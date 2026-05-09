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
	uploadDir      string
}

// NewMessageHandler creates a new MessageHandler.
func NewMessageHandler(ms *services.MessageService, es *services.EmailService, ps *services.ProfileService, broker *services.EventBroker, uploadDir string) *MessageHandler {
	return &MessageHandler{msgService: ms, emailService: es, profileService: ps, broker: broker, uploadDir: uploadDir}
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

// Reply sends an email reply to the message sender [protected].
func (h *MessageHandler) Reply(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid message ID", nil)
		return
	}

	// Parse as multipart form so we can handle file attachments
	if err := c.Request.ParseMultipartForm(50 << 20); err != nil {
		// Fall back to JSON if no files attached
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

	// Collect attachments. The dashboard pre-uploads files via /upload BEFORE
	// hitting this endpoint and passes their URLs in `attachment_urls`. Reading
	// those files from disk here avoids resending the binary over the wire (which
	// previously made every reply with attachments take ~2× as long and caused
	// "NetworkError" timeouts on slow connections). The multipart binary loop
	// below is kept as a fallback for clients that don't pre-upload.
	var attachments []services.Attachment
	if attUrlsJSON != "" {
		var meta []struct {
			Name string `json:"name"`
			URL  string `json:"url"`
			Type string `json:"type"`
		}
		if err := json.Unmarshal([]byte(attUrlsJSON), &meta); err == nil {
			for _, m := range meta {
				if m.Type == "link" || m.URL == "" {
					continue // links live in the body text, not as MIME attachments
				}
				if !strings.HasPrefix(m.URL, "/uploads/") {
					continue // only serve our own files
				}
				rel := strings.TrimPrefix(m.URL, "/uploads/")
				full, err := filepath.Abs(filepath.Join(h.uploadDir, rel))
				if err != nil {
					continue
				}
				// Make sure the resolved path is still inside uploadDir (no traversal)
				absUploadDir, _ := filepath.Abs(h.uploadDir)
				if !strings.HasPrefix(full, absUploadDir) {
					continue
				}
				data, err := os.ReadFile(full)
				if err != nil {
					continue
				}
				ct := mime.TypeByExtension(filepath.Ext(m.Name))
				if ct == "" {
					ct = "application/octet-stream"
				}
				attachments = append(attachments, services.Attachment{
					Filename:    m.Name,
					ContentType: ct,
					Data:        data,
				})
			}
		}
	}

	// Fallback: if no pre-uploaded URLs were provided, accept binary files from
	// the multipart form directly (legacy clients).
	if len(attachments) == 0 && c.Request.MultipartForm != nil {
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
				ct := fh.Header.Get("Content-Type")
				if ct == "" {
					ct = mime.TypeByExtension(fh.Filename)
				}
				if ct == "" {
					ct = "application/octet-stream"
				}
				attachments = append(attachments, services.Attachment{
					Filename:    fh.Filename,
					ContentType: ct,
					Data:        data,
				})
			}
		}
	}

	// Alias so code below uses same names
	body := struct {
		ReplyText string
		FromEmail string
	}{ReplyText: replyText, FromEmail: fromEmail}

	// Fetch the original message so we can reply to the right person
	msg, err := h.msgService.GetByID(uint(id))
	if err != nil {
		utils.NotFound(c, "Message not found")
		return
	}

	subject := "Re: " + msg.Type + " — Imtiaz Portfolio"
	fullBody := body.ReplyText + "\n\n---\nOriginal message from " + msg.Name + ":\n" + msg.Content

	// Use SMTP credentials from DB profile if set — overrides .env
	emailSvc := h.emailService
	if profile, err := h.profileService.Get(); err == nil && profile.SMTPUser != "" && profile.SMTPPass != "" {
		emailSvc = services.NewEmailService(
			h.emailService.Host(),
			h.emailService.Port(),
			profile.SMTPUser,
			profile.SMTPPass,
		)
	}

	if err := emailSvc.Send(msg.Name, msg.Email, body.FromEmail, subject, fullBody, attachments); err != nil {
		utils.InternalError(c, "Failed to send email: "+err.Error())
		return
	}

	// Use pre-uploaded file URLs (parsed at the top) for chat-history rendering,
	// otherwise fall back to bare filenames so legacy clients still work.
	attachmentNames := attUrlsJSON
	if attachmentNames == "" {
		var filenames []string
		for _, att := range attachments {
			filenames = append(filenames, att.Filename)
		}
		attachmentNames = strings.Join(filenames, ",")
	}

	// Save the reply as a new record so full history is preserved
	_ = h.msgService.CreateReply(&models.MessageReply{
		MessageID:   uint(id),
		ReplyText:   body.ReplyText,
		FromEmail:   body.FromEmail,
		Attachments: attachmentNames,
	})

	// Also update the message flags
	_ = h.msgService.MarkAsReplied(uint(id), body.ReplyText, attachmentNames)
	_ = h.msgService.MarkAsRead(uint(id))

	utils.Success(c, "Reply sent successfully", nil)
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
