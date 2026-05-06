// ============================================================
// FILE: internal/services/profile_service.go
// WHAT IT IS:     Portfolio profile read/update business logic
// WHY IT EXISTS:  The profile is the single editable record that
//                 powers the hero, about, and contact sections.
//                 This service ensures only one profile ever exists.
// ENDPOINTS:      Called by profile_handler.go
// DEPENDS ON:     models/profile.go, gorm
// IF REMOVED:     Profile cannot be read or updated
// LAST UPDATED:   2026-05-07 — initial creation
// ============================================================

package services

import (
	"imtiaz-portfolio/internal/models"
	"gorm.io/gorm"
)

// ProfileService handles database operations for the single profile record.
type ProfileService struct {
	db *gorm.DB
}

// NewProfileService creates a new ProfileService.
func NewProfileService(db *gorm.DB) *ProfileService {
	return &ProfileService{db: db}
}

// Get returns the single profile record (always ID=1).
// If no profile exists yet, returns an empty profile struct.
func (s *ProfileService) Get() (*models.Profile, error) {
	var profile models.Profile
	result := s.db.First(&profile)
	if result.Error == gorm.ErrRecordNotFound {
		// Return empty profile rather than error — seed may not have run yet
		return &models.Profile{}, nil
	}
	return &profile, result.Error
}

// Update saves changes to the existing profile record.
// Uses Save() which updates ALL fields, not just changed ones,
// so passing the full profile struct is required.
func (s *ProfileService) Update(updates *models.Profile) (*models.Profile, error) {
	var profile models.Profile
	s.db.First(&profile)
	if profile.ID == 0 {
		// No profile exists yet — create one
		if err := s.db.Create(updates).Error; err != nil {
			return nil, err
		}
		return updates, nil
	}
	updates.ID = profile.ID
	if err := s.db.Save(updates).Error; err != nil {
		return nil, err
	}
	return updates, nil
}
