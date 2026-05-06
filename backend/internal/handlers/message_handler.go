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
	"imtiaz-portfolio/internal/models"
	"imtiaz-portfolio/internal/services"
	"imtiaz-portfolio/internal/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// MessageHandler holds the message service dependency.
type MessageHandler struct {
	msgService *services.MessageService
}

// NewMessageHandler creates a new MessageHandler.
func NewMessageHandler(ms *services.MessageService) *MessageHandler {
	return &MessageHandler{msgService: ms}
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
