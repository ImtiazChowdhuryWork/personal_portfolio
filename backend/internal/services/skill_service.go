// ============================================================
// FILE: internal/services/skill_service.go
// WHAT IT IS:     Skill CRUD business logic
// WHY IT EXISTS:  Database operations for the Tech Stack section
// ENDPOINTS:      Called by skill_handler.go
// DEPENDS ON:     models/skill.go, gorm
// IF REMOVED:     Skill handler has no database operations
// LAST UPDATED:   2026-05-07 — initial creation
// ============================================================

package services

import (
	"errors"
	"imtiaz-portfolio/internal/models"
	"gorm.io/gorm"
)

// SkillService handles all database operations for skills.
type SkillService struct {
	db *gorm.DB
}

// NewSkillService creates a new SkillService.
func NewSkillService(db *gorm.DB) *SkillService {
	return &SkillService{db: db}
}

// GetAll returns all skills ordered by category then sort order.
func (s *SkillService) GetAll() ([]models.Skill, error) {
	var skills []models.Skill
	result := s.db.Order("category ASC, sort_order ASC").Find(&skills)
	return skills, result.Error
}

// Create inserts a new skill into the database.
func (s *SkillService) Create(skill *models.Skill) error {
	return s.db.Create(skill).Error
}

// Update modifies an existing skill by ID.
func (s *SkillService) Update(id uint, updates *models.Skill) (*models.Skill, error) {
	var skill models.Skill
	if err := s.db.First(&skill, id).Error; err != nil {
		return nil, errors.New("skill not found")
	}
	s.db.Model(&skill).Updates(updates)
	return &skill, nil
}

// Delete soft-deletes a skill by ID.
func (s *SkillService) Delete(id uint) error {
	result := s.db.Delete(&models.Skill{}, id)
	if result.RowsAffected == 0 {
		return errors.New("skill not found")
	}
	return result.Error
}
