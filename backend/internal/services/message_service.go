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

// CountUnread returns the number of unread messages for the dashboard badge.
func (s *MessageService) CountUnread() (int64, error) {
	var count int64
	result := s.db.Model(&models.Message{}).Where("is_read = ?", false).Count(&count)
	return count, result.Error
}
