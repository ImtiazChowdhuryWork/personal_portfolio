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
	"strings"

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

// UpdateField updates a single column on the profile without touching other fields.
func (s *ProfileService) UpdateField(column string, value interface{}) error {
	return s.db.Model(&models.Profile{}).Where("id = 1").Update(column, value).Error
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

	// Strip spaces from App Password
	updates.SMTPPass = strings.ReplaceAll(updates.SMTPPass, " ", "")

	// For every string field, if the incoming value is empty preserve the existing one.
	// This prevents accidentally wiping data when only some fields are submitted.
	preserve := func(incoming, existing string) string {
		if incoming == "" {
			return existing
		}
		return incoming
	}

	updates.FullName        = preserve(updates.FullName, profile.FullName)
	updates.Title           = preserve(updates.Title, profile.Title)
	updates.Tagline         = preserve(updates.Tagline, profile.Tagline)
	updates.Bio             = preserve(updates.Bio, profile.Bio)
	updates.ShortBio        = preserve(updates.ShortBio, profile.ShortBio)
	updates.Email           = preserve(updates.Email, profile.Email)
	updates.Phone           = preserve(updates.Phone, profile.Phone)
	updates.WhatsApp        = preserve(updates.WhatsApp, profile.WhatsApp)
	updates.Location        = preserve(updates.Location, profile.Location)
	updates.GitHub          = preserve(updates.GitHub, profile.GitHub)
	updates.LinkedIn        = preserve(updates.LinkedIn, profile.LinkedIn)
	updates.Twitter         = preserve(updates.Twitter, profile.Twitter)
	updates.Instagram       = preserve(updates.Instagram, profile.Instagram)
	updates.ProfilePhoto    = preserve(updates.ProfilePhoto, profile.ProfilePhoto)
	updates.AboutPhoto      = preserve(updates.AboutPhoto, profile.AboutPhoto)
	updates.CVFile          = preserve(updates.CVFile, profile.CVFile)
	updates.Availability    = preserve(updates.Availability, profile.Availability)
	updates.YearsExperience = preserve(updates.YearsExperience, profile.YearsExperience)
	updates.AppsShipped     = preserve(updates.AppsShipped, profile.AppsShipped)
	updates.SMTPUser        = preserve(updates.SMTPUser, profile.SMTPUser)
	updates.SMTPPass        = preserve(updates.SMTPPass, profile.SMTPPass)
	updates.MetaTitle       = preserve(updates.MetaTitle, profile.MetaTitle)
	updates.MetaDescription = preserve(updates.MetaDescription, profile.MetaDescription)

	if err := s.db.Save(updates).Error; err != nil {
		return nil, err
	}
	return updates, nil
}
