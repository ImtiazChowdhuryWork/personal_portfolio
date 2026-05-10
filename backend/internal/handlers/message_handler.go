// ============================================================
// FILE: internal/handlers/message_handler.go
// WHAT IT IS:     HTTP handlers for the contact messages API
// ENDPOINTS:      POST   /api/v1/messages          (public — contact form)
//                 GET    /api/v1/messages          [protected — dashboard]
//                 PUT    /api/v1/messages/:id      [protected — mark read]
//                 DELETE /api/v1/messages/:id      [protected]
// DEPENDS ON:     services/message_service.go, utils/response.go
// LAST UPDATED:   2026-05-07 — initial creation
// ============================================================

package handlers

import (
	"encoding/json"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"imtiaz-portfolio/internal/models"
	"imtiaz-portfolio/internal/services"
	"imtiaz-portfolio/internal/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// MessageHandler holds the message service dependency.
type MessageHandler struct {
	msgService     *services.MessageService
	emailService   *services.EmailService
	profileService *services.ProfileService
	broker         *services.EventBroker
	replyQueue     *services.ReplyQueue
	uploadDir      string
}

// NewMessageHandler creates a new MessageHandler.
func NewMessageHandler(ms *services.MessageService, es *services.EmailService, ps *services.ProfileService, broker *services.EventBroker, queue *services.ReplyQueue, uploadDir string) *MessageHandler {
	return &MessageHandler{msgService: ms, emailService: es, profileService: ps, broker: broker, replyQueue: queue, uploadDir: uploadDir}
}

// GetAll returns all contact messages — protected, dashboard only.
func (h *MessageHandler) GetAll(c *gin.Context) {
	messages, err := h.msgService.GetAll()
	if err != nil {
		utils.InternalError(c, "Failed to fetch messages")
		return
	}
	utils.Success(c, "Messages fetched successfully", messages)
}

// Create handles the public contact form submission.
// This endpoint is NOT protected — anyone can submit a contact form.
func (h *MessageHandler) Create(c *gin.Context) {
	var msg models.Message
	if err := c.ShouldBindJSON(&msg); err != nil {
		utils.BadRequest(c, "Invalid message data", err.Error())
		return
	}

	// Validate required fields from the contact form
	if utils.IsEmpty(msg.Name) {
		utils.BadRequest(c, "Your name is required", nil)
		return
	}
	if !utils.IsValidEmail(msg.Email) {
		utils.BadRequest(c, "A valid email address is required", nil)
		return
	}
	if utils.IsEmpty(msg.Content) {
		utils.BadRequest(c, "Please write a message", nil)
		return
	}

	// Record the sender's IP for spam filtering purposes
	msg.IPAddress = c.ClientIP()

	if err := h.msgService.Create(&msg); err != nil {
		utils.InternalError(c, "Failed to save your message — please try again")
		return
	}
	// Push a real-time notification to every connected dashboard so the
	// unread badge updates without waiting for the next poll.
	if h.broker != nil {
		h.broker.Publish(services.Event{
			Type: "message.created",
			Data: map[string]interface{}{
				"id":         msg.ID,
				"name":       msg.Name,
				"email":      msg.Email,
				"type":       msg.Type,
				"created_at": msg.CreatedAt,
			},
		})
	}
	utils.Created(c, "Message sent successfully! I'll get back to you soon.", msg)
}

// MarkAsRead marks a message as read — called from the dashboard [protected].
func (h *MessageHandler) MarkAsRead(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid message ID", nil)
		return
	}
	if err := h.msgService.MarkAsRead(uint(id)); err != nil {
		utils.NotFound(c, err.Error())
		return
	}
	utils.Success(c, "Message marked as read", nil)
}

// Reply enqueues an email reply for asynchronous SMTP delivery [protected].
//
// The handler used to call SMTP synchronously, which blocked the admin's
// browser for 10–150s on replies with large attachments. Now it:
//   1. Persists the reply with delivery_status="pending"
//   2. Marks the original message as replied + read
//   3. Pushes a job onto the in-memory queue
//   4. Returns 200 OK with the new reply ID — typically in <100 ms
//
// A worker goroutine picks the job up, calls SMTP, and updates the row to
// "sent" or "failed". The dashboard renders the bubble immediately and
// flips its tick when a reply.sent / reply.failed SSE event arrives.
func (h *MessageHandler) Reply(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid message ID", nil)
		return
	}

	// Parse as multipart form so we can handle file attachments
	if err := c.Request.ParseMultipartForm(50 << 20); err != nil {
		utils.BadRequest(c, "Invalid form data", nil)
		return
	}
	replyText := c.Request.FormValue("reply_text")
	fromEmail := c.Request.FormValue("from_email")

	if replyText == "" {
		utils.BadRequest(c, "Reply text is required", nil)
		return
	}

	attUrlsJSON := c.Request.FormValue("attachment_urls")

	// Legacy clients that don't pre-upload files still POST binary multipart
	// parts. Persist them to /uploads/replies/<ts>_<name> here so the queue
	// worker can read them back from disk like a normal pre-upload.
	if attUrlsJSON == "" && c.Request.MultipartForm != nil {
		stored := h.persistLegacyMultipart(c)
		if len(stored) > 0 {
			b, _ := json.Marshal(stored)
			attUrlsJSON = string(b)
		}
	}

	// Fetch the original message so the worker can build the right To: + body.
	msg, err := h.msgService.GetByID(uint(id))
	if err != nil {
		utils.NotFound(c, "Message not found")
		return
	}

	subject := "Re: " + msg.Type + " — Imtiaz Portfolio"
	fullBody := replyText + "\n\n---\nOriginal message from " + msg.Name + ":\n" + msg.Content

	// Persist the reply row up front — it owns the delivery state. Worker
	// updates this same row when SMTP returns.
	reply := &models.MessageReply{
		MessageID:      uint(id),
		ReplyText:      replyText,
		FromEmail:      fromEmail,
		Attachments:    attUrlsJSON,
		DeliveryStatus: "pending",
	}
	if err := h.msgService.CreateReply(reply); err != nil {
		utils.InternalError(c, "Failed to save reply")
		return
	}

	// Update inbox flags so the conversation list reflects "replied" without
	// waiting for SMTP. If delivery later fails, the bubble's ⚠ icon makes
	// that obvious; the message itself is still legitimately answered.
	_ = h.msgService.MarkAsReplied(uint(id), replyText, attUrlsJSON)
	_ = h.msgService.MarkAsRead(uint(id))

	// Hand off to the worker. From here the admin sees the bubble appear
	// instantly; the worker takes however long it takes.
	if h.replyQueue != nil {
		h.replyQueue.Enqueue(services.ReplyJob{
			ReplyID:        reply.ID,
			ToName:         msg.Name,
			ToEmail:        msg.Email,
			FromEmail:      fromEmail,
			Subject:        subject,
			Body:           fullBody,
			AttachmentURLs: attUrlsJSON,
		})
	} else {
		// No queue wired — should never happen in production. Leave the row
		// pending and surface so the operator notices.
		_ = h.msgService.MarkReplyFailed(reply.ID, "reply queue not initialised")
	}

	utils.Created(c, "Reply queued for delivery", gin.H{
		"reply_id":        reply.ID,
		"delivery_status": "pending",
	})
}

// persistLegacyMultipart writes any binary parts on the request to
// /uploads/replies/<unix-ms>_<filename> and returns metadata in the same
// shape the worker expects in attachment_urls. Lets old client builds keep
// working without changing the queue.
func (h *MessageHandler) persistLegacyMultipart(c *gin.Context) []map[string]interface{} {
	if c.Request.MultipartForm == nil {
		return nil
	}
	dir := filepath.Join(h.uploadDir, "replies")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil
	}
	var out []map[string]interface{}
	for _, fileHeaders := range c.Request.MultipartForm.File {
		for _, fh := range fileHeaders {
			f, err := fh.Open()
			if err != nil {
				continue
			}
			data, err := io.ReadAll(f)
			f.Close()
			if err != nil {
				continue
			}
			safeName := strings.ReplaceAll(fh.Filename, "/", "_")
			safeName = strings.ReplaceAll(safeName, "\\", "_")
			full := filepath.Join(dir, safeName)
			if err := os.WriteFile(full, data, 0o644); err != nil {
				continue
			}
			ct := fh.Header.Get("Content-Type")
			if ct == "" {
				ct = mime.TypeByExtension(filepath.Ext(fh.Filename))
			}
			out = append(out, map[string]interface{}{
				"name": fh.Filename,
				"url":  "/uploads/replies/" + safeName,
				"type": ct,
				"size": fh.Size,
			})
		}
	}
	return out
}

// RetryReply re-enqueues a previously-failed reply [protected].
// Looks up the row by reply ID, flips it back to "pending", and pushes a
// fresh job onto the queue. Frontend retry button hits this.
func (h *MessageHandler) RetryReply(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid reply ID", nil)
		return
	}
	reply, err := h.msgService.GetReplyByID(uint(id))
	if err != nil {
		utils.NotFound(c, "Reply not found")
		return
	}
	msg, err := h.msgService.GetByID(reply.MessageID)
	if err != nil {
		utils.NotFound(c, "Original message not found")
		return
	}
	if err := h.msgService.MarkReplyPending(uint(id)); err != nil {
		utils.InternalError(c, "Failed to reset delivery state")
		return
	}

	subject := "Re: " + msg.Type + " — Imtiaz Portfolio"
	fullBody := reply.ReplyText + "\n\n---\nOriginal message from " + msg.Name + ":\n" + msg.Content

	if h.replyQueue != nil {
		h.replyQueue.Enqueue(services.ReplyJob{
			ReplyID:        reply.ID,
			ToName:         msg.Name,
			ToEmail:        msg.Email,
			FromEmail:      reply.FromEmail,
			Subject:        subject,
			Body:           fullBody,
			AttachmentURLs: reply.Attachments,
		})
	}

	utils.Success(c, "Retry queued", gin.H{
		"reply_id":        reply.ID,
		"delivery_status": "pending",
	})
}

// GetReplies returns all replies for a message [protected].
func (h *MessageHandler) GetReplies(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid message ID", nil)
		return
	}
	replies, err := h.msgService.GetReplies(uint(id))
	if err != nil {
		utils.InternalError(c, "Failed to fetch replies")
		return
	}
	utils.Success(c, "Replies fetched", replies)
}

// Delete removes a message [protected].
func (h *MessageHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid message ID", nil)
		return
	}
	if err := h.msgService.Delete(uint(id)); err != nil {
		utils.NotFound(c, err.Error())
		return
	}
	utils.Success(c, "Message deleted successfully", nil)
}
