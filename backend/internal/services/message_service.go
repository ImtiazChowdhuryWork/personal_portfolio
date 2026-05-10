// ============================================================
// FILE: internal/services/message_service.go
// WHAT IT IS:     Contact message CRUD business logic
// WHY IT EXISTS:  Database operations for the contact form submissions
// ENDPOINTS:      Called by message_handler.go
// DEPENDS ON:     models/message.go, gorm
// IF REMOVED:     Contact form cannot save or retrieve messages
// LAST UPDATED:   2026-05-07 — initial creation
// ============================================================

package services

import (
	"errors"
	"time"

	"imtiaz-portfolio/internal/models"

	"gorm.io/gorm"
)

// MessageService handles all database operations for contact messages.
type MessageService struct {
	db *gorm.DB
}

// NewMessageService creates a new MessageService.
func NewMessageService(db *gorm.DB) *MessageService {
	return &MessageService{db: db}
}

// GetAll returns all messages ordered by newest first (for the dashboard inbox).
func (s *MessageService) GetAll() ([]models.Message, error) {
	var messages []models.Message
	result := s.db.Order("created_at DESC").Find(&messages)
	return messages, result.Error
}

// Create saves a new contact form message to the database.
func (s *MessageService) Create(msg *models.Message) error {
	return s.db.Create(msg).Error
}

// MarkAsRead sets IsRead=true on a message so the dashboard badge decrements.
func (s *MessageService) MarkAsRead(id uint) error {
	result := s.db.Model(&models.Message{}).Where("id = ?", id).Update("is_read", true)
	if result.RowsAffected == 0 {
		return errors.New("message not found")
	}
	return result.Error
}

// Delete soft-deletes a message by ID.
func (s *MessageService) Delete(id uint) error {
	result := s.db.Delete(&models.Message{}, id)
	if result.RowsAffected == 0 {
		return errors.New("message not found")
	}
	return result.Error
}

// GetByID returns a single message by ID.
func (s *MessageService) GetByID(id uint) (*models.Message, error) {
	var msg models.Message
	result := s.db.First(&msg, id)
	if result.Error != nil {
		return nil, errors.New("message not found")
	}
	return &msg, nil
}

// MarkAsReplied sets IsReplied=true, saves the reply text and attachment filenames.
func (s *MessageService) MarkAsReplied(id uint, replyText, attachmentNames string) error {
	return s.db.Model(&models.Message{}).Where("id = ?", id).Updates(map[string]interface{}{
		"is_replied":        true,
		"reply_text":        replyText,
		"reply_attachments": attachmentNames,
	}).Error
}

// CreateReply saves a new reply record for a message.
func (s *MessageService) CreateReply(reply *models.MessageReply) error {
	return s.db.Create(reply).Error
}

// GetReplies returns all replies for a message ordered oldest first.
func (s *MessageService) GetReplies(messageID uint) ([]models.MessageReply, error) {
	var replies []models.MessageReply
	err := s.db.Where("message_id = ?", messageID).Order("created_at ASC").Find(&replies).Error
	return replies, err
}

// GetReplyByID returns one reply row — used by the retry endpoint to
// rebuild a worker job from a previously-failed delivery.
func (s *MessageService) GetReplyByID(id uint) (*models.MessageReply, error) {
	var reply models.MessageReply
	if err := s.db.First(&reply, id).Error; err != nil {
		return nil, errors.New("reply not found")
	}
	return &reply, nil
}

// MarkReplySent flips the row to delivery_status="sent" + sets DeliveredAt.
// Called by the queue worker after SMTP returns 250 OK.
func (s *MessageService) MarkReplySent(id uint) error {
	now := time.Now()
	return s.db.Model(&models.MessageReply{}).Where("id = ?", id).Updates(map[string]interface{}{
		"delivery_status": "sent",
		"delivery_error":  "",
		"delivered_at":    &now,
	}).Error
}

// MarkReplyFailed flips the row to delivery_status="failed" with the wrapped
// SMTP error so the dashboard can show ⚠ + the reason on hover.
func (s *MessageService) MarkReplyFailed(id uint, errMsg string) error {
	return s.db.Model(&models.MessageReply{}).Where("id = ?", id).Updates(map[string]interface{}{
		"delivery_status": "failed",
		"delivery_error":  errMsg,
	}).Error
}

// MarkReplyPending resets a row to pending — used when retrying a failed
// reply so the UI tick animation is consistent with a fresh send.
func (s *MessageService) MarkReplyPending(id uint) error {
	return s.db.Model(&models.MessageReply{}).Where("id = ?", id).Updates(map[string]interface{}{
		"delivery_status": "pending",
		"delivery_error":  "",
	}).Error
}

// GetPendingReplies returns every reply still in "pending" — used at server
// startup to recover jobs that were in-flight when the process exited.
func (s *MessageService) GetPendingReplies() ([]models.MessageReply, error) {
	var replies []models.MessageReply
	err := s.db.Where("delivery_status = ?", "pending").Order("created_at ASC").Find(&replies).Error
	return replies, err
}

// CountUnread returns the number of unread messages for the dashboard badge.
func (s *MessageService) CountUnread() (int64, error) {
	var count int64
	result := s.db.Model(&models.Message{}).Where("is_read = ?", false).Count(&count)
	return count, result.Error
}
