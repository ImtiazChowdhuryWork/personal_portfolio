// ============================================================
// FILE: internal/services/experience_service.go
// WHAT IT IS:     Work experience CRUD business logic
// WHY IT EXISTS:  Database operations for the Experience Timeline section
// ENDPOINTS:      Called by experience_handler.go
// DEPENDS ON:     models/experience.go, gorm
// IF REMOVED:     Experience handler has no database operations
// LAST UPDATED:   2026-05-07 — initial creation
// ============================================================

package services

import (
	"errors"
	"imtiaz-portfolio/internal/models"
	"gorm.io/gorm"
)

// ExperienceService handles all database operations for work experience.
type ExperienceService struct {
	db *gorm.DB
}

// NewExperienceService creates a new ExperienceService.
func NewExperienceService(db *gorm.DB) *ExperienceService {
	return &ExperienceService{db: db}
}

// GetAll returns all experience entries ordered by sort order (most recent first).
func (s *ExperienceService) GetAll() ([]models.Experience, error) {
	var experiences []models.Experience
	result := s.db.Order("sort_order ASC, created_at DESC").Find(&experiences)
	return experiences, result.Error
}

// Create inserts a new experience entry.
func (s *ExperienceService) Create(exp *models.Experience) error {
	return s.db.Create(exp).Error
}

// Update modifies an existing experience entry by ID.
func (s *ExperienceService) Update(id uint, updates *models.Experience) (*models.Experience, error) {
	var exp models.Experience
	if err := s.db.First(&exp, id).Error; err != nil {
		return nil, errors.New("experience not found")
	}
	s.db.Model(&exp).Updates(updates)
	return &exp, nil
}

// Delete soft-deletes an experience entry by ID.
func (s *ExperienceService) Delete(id uint) error {
	result := s.db.Delete(&models.Experience{}, id)
	if result.RowsAffected == 0 {
		return errors.New("experience not found")
	}
	return result.Error
}
