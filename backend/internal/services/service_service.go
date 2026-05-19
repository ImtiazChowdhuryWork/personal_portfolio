// ============================================================
// FILE: internal/services/service_service.go
// WHAT IT IS:     Service/offering CRUD business logic
// ENDPOINTS:      Called by service_handler.go
// DEPENDS ON:     models/service.go, gorm
// LAST UPDATED:   2026-05-19 — renormalize after every mutation
// ============================================================

package services

import (
	"errors"
	"imtiaz-portfolio/internal/models"

	"gorm.io/gorm"
)

type ServiceService struct {
	db *gorm.DB
}

func NewServiceService(db *gorm.DB) *ServiceService {
	return &ServiceService{db: db}
}

// GetAll returns all active services ordered by sort_order.
func (s *ServiceService) GetAll() ([]models.Service, error) {
	var items []models.Service
	result := s.db.Where("active = ?", true).Order("sort_order ASC, id ASC").Find(&items)
	return items, result.Error
}

// GetAllAdmin returns all services (including inactive) for the dashboard.
func (s *ServiceService) GetAllAdmin() ([]models.Service, error) {
	var items []models.Service
	result := s.db.Order("sort_order ASC, id ASC").Find(&items)
	return items, result.Error
}

// renormalize re-sequences all non-deleted services to 1, 2, 3 …
// Called after every create / update / delete so sort_orders stay
// gap-free and duplicate-free regardless of what the caller sent.
func (s *ServiceService) renormalize() {
	var items []models.Service
	s.db.Order("sort_order ASC, id ASC").Find(&items)
	for i, item := range items {
		s.db.Exec("UPDATE services SET sort_order = ? WHERE id = ?", i+1, item.ID)
	}
}

func (s *ServiceService) Create(service *models.Service) error {
	// Make room at the desired position
	s.db.Exec(
		"UPDATE services SET sort_order = sort_order + 1 WHERE sort_order >= ? AND deleted_at IS NULL",
		service.SortOrder,
	)
	if err := s.db.Create(service).Error; err != nil {
		return err
	}
	// Normalize to close any gaps from previous bad state
	s.renormalize()
	return nil
}

func (s *ServiceService) Update(id uint, updates *models.Service) (*models.Service, error) {
	var service models.Service
	if err := s.db.First(&service, id).Error; err != nil {
		return nil, errors.New("service not found")
	}

	oldOrder := service.SortOrder
	newOrder := updates.SortOrder

	if newOrder != oldOrder {
		if newOrder < oldOrder {
			// Moving to an earlier position: push the services in between down
			s.db.Exec(
				"UPDATE services SET sort_order = sort_order + 1 WHERE sort_order >= ? AND sort_order < ? AND id != ? AND deleted_at IS NULL",
				newOrder, oldOrder, id,
			)
		} else {
			// Moving to a later position: pull the services in between up
			s.db.Exec(
				"UPDATE services SET sort_order = sort_order - 1 WHERE sort_order > ? AND sort_order <= ? AND id != ? AND deleted_at IS NULL",
				oldOrder, newOrder, id,
			)
		}
	}

	if err := s.db.Model(&service).Select(
		"icon", "title", "description", "sort_order", "active",
	).Updates(updates).Error; err != nil {
		return nil, err
	}

	// Normalize to fix any edge cases (e.g. user typed order 99 on a 3-item list)
	s.renormalize()

	s.db.First(&service, id)
	return &service, nil
}

func (s *ServiceService) Delete(id uint) error {
	var service models.Service
	if err := s.db.First(&service, id).Error; err != nil {
		return errors.New("service not found")
	}
	result := s.db.Delete(&models.Service{}, id)
	if result.RowsAffected == 0 {
		return errors.New("service not found")
	}
	if result.Error != nil {
		return result.Error
	}
	// Close the gap then normalize the full sequence
	s.db.Exec(
		"UPDATE services SET sort_order = sort_order - 1 WHERE sort_order > ? AND deleted_at IS NULL",
		service.SortOrder,
	)
	s.renormalize()
	return nil
}
