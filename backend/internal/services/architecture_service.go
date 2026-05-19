// ============================================================
// FILE: internal/services/architecture_service.go
// WHAT IT IS:     Architecture CRUD business logic
// LAST UPDATED:   2026-05-19 — initial creation
// ============================================================

package services

import (
	"errors"
	"imtiaz-portfolio/internal/models"

	"gorm.io/gorm"
)

type ArchitectureService struct {
	db *gorm.DB
}

func NewArchitectureService(db *gorm.DB) *ArchitectureService {
	return &ArchitectureService{db: db}
}

func (s *ArchitectureService) GetAll() ([]models.Architecture, error) {
	var items []models.Architecture
	result := s.db.Where("active = ?", true).Order("sort_order ASC, id ASC").Find(&items)
	return items, result.Error
}

func (s *ArchitectureService) GetAllAdmin() ([]models.Architecture, error) {
	var items []models.Architecture
	result := s.db.Order("sort_order ASC, id ASC").Find(&items)
	return items, result.Error
}

func (s *ArchitectureService) renormalize() {
	var items []models.Architecture
	s.db.Order("sort_order ASC, id ASC").Find(&items)
	for i, item := range items {
		s.db.Exec("UPDATE architectures SET sort_order = ? WHERE id = ?", i+1, item.ID)
	}
}

func (s *ArchitectureService) Create(arch *models.Architecture) error {
	s.db.Exec(
		"UPDATE architectures SET sort_order = sort_order + 1 WHERE sort_order >= ? AND deleted_at IS NULL",
		arch.SortOrder,
	)
	if err := s.db.Create(arch).Error; err != nil {
		return err
	}
	s.renormalize()
	return nil
}

func (s *ArchitectureService) Update(id uint, updates *models.Architecture) (*models.Architecture, error) {
	var arch models.Architecture
	if err := s.db.First(&arch, id).Error; err != nil {
		return nil, errors.New("architecture not found")
	}

	oldOrder := arch.SortOrder
	newOrder := updates.SortOrder

	if newOrder != oldOrder {
		if newOrder < oldOrder {
			s.db.Exec(
				"UPDATE architectures SET sort_order = sort_order + 1 WHERE sort_order >= ? AND sort_order < ? AND id != ? AND deleted_at IS NULL",
				newOrder, oldOrder, id,
			)
		} else {
			s.db.Exec(
				"UPDATE architectures SET sort_order = sort_order - 1 WHERE sort_order > ? AND sort_order <= ? AND id != ? AND deleted_at IS NULL",
				oldOrder, newOrder, id,
			)
		}
	}

	if err := s.db.Model(&arch).Select(
		"number", "name", "diagram", "description", "projects", "sort_order", "active",
	).Updates(updates).Error; err != nil {
		return nil, err
	}
	s.renormalize()
	s.db.First(&arch, id)
	return &arch, nil
}

func (s *ArchitectureService) Delete(id uint) error {
	var arch models.Architecture
	if err := s.db.First(&arch, id).Error; err != nil {
		return errors.New("architecture not found")
	}
	result := s.db.Delete(&models.Architecture{}, id)
	if result.RowsAffected == 0 {
		return errors.New("architecture not found")
	}
	if result.Error != nil {
		return result.Error
	}
	s.db.Exec(
		"UPDATE architectures SET sort_order = sort_order - 1 WHERE sort_order > ? AND deleted_at IS NULL",
		arch.SortOrder,
	)
	s.renormalize()
	return nil
}
