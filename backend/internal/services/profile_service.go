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
	"time"

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

// UpdateFields updates a specific set of columns on the profile in one query.
// Using a map (instead of a struct) ensures GORM writes every key — including
// empty strings — so values can be cleared from the dashboard.
func (s *ProfileService) UpdateFields(fields map[string]interface{}) error {
	return s.db.Model(&models.Profile{}).Where("id = 1").Updates(fields).Error
}

// AddPasswordHistory records a Gmail App Password change in the history table.
// Skips inserting when the password matches the most recent entry, so re-saving
// the form without changing the password doesn't create duplicate rows.
func (s *ProfileService) AddPasswordHistory(gmailAddress, appPassword string) error {
	if appPassword == "" {
		return nil
	}
	var latest models.MailPasswordHistory
	s.db.Order("created_at DESC").First(&latest)
	if latest.AppPassword == appPassword && latest.GmailAddress == gmailAddress {
		return nil
	}
	entry := models.MailPasswordHistory{
		GmailAddress: gmailAddress,
		AppPassword:  appPassword,
	}
	return s.db.Create(&entry).Error
}

// GetPasswordHistory returns every saved Gmail App Password, newest first.
func (s *ProfileService) GetPasswordHistory() ([]models.MailPasswordHistory, error) {
	var history []models.MailPasswordHistory
	err := s.db.Order("created_at DESC").Find(&history).Error
	return history, err
}

// GetPasswordHistoryByID fetches a single history row by primary key.
func (s *ProfileService) GetPasswordHistoryByID(id uint) (*models.MailPasswordHistory, error) {
	var row models.MailPasswordHistory
	if err := s.db.First(&row, id).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

// DeletePasswordHistory removes a single entry from the history table.
func (s *ProfileService) DeletePasswordHistory(id uint) error {
	return s.db.Delete(&models.MailPasswordHistory{}, id).Error
}

// DeleteMailAccount removes every history row that matches the given Gmail
// address. Used to wipe an entire account from the dashboard's Available
// Accounts panel in a single action.
func (s *ProfileService) DeleteMailAccount(email string) error {
	return s.db.Where("gmail_address = ?", email).Delete(&models.MailPasswordHistory{}).Error
}

// HideMailAccount marks an email as hidden from the Available Accounts panel
// without touching its history rows. Idempotent — calling it twice is a no-op.
func (s *ProfileService) HideMailAccount(email string) error {
	row := models.HiddenMailAccount{Email: email, HiddenAt: time.Now()}
	return s.db.Where("email = ?", email).FirstOrCreate(&row, row).Error
}

// UnhideMailAccount restores a hidden email so it shows up in Available Accounts again.
func (s *ProfileService) UnhideMailAccount(email string) error {
	return s.db.Where("email = ?", email).Delete(&models.HiddenMailAccount{}).Error
}

// GetHiddenMailEmails returns the set of emails currently marked hidden.
func (s *ProfileService) GetHiddenMailEmails() ([]string, error) {
	var emails []string
	err := s.db.Model(&models.HiddenMailAccount{}).Pluck("email", &emails).Error
	return emails, err
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
	updates.GitHubUsername    = preserve(updates.GitHubUsername, profile.GitHubUsername)
	updates.GitHubRepos       = preserve(updates.GitHubRepos, profile.GitHubRepos)
	updates.GitHubCommits     = preserve(updates.GitHubCommits, profile.GitHubCommits)
	updates.GitHubTopLanguage = preserve(updates.GitHubTopLanguage, profile.GitHubTopLanguage)
	updates.GitHubYearsActive = preserve(updates.GitHubYearsActive, profile.GitHubYearsActive)
	updates.MetaTitle       = preserve(updates.MetaTitle, profile.MetaTitle)
	updates.MetaDescription = preserve(updates.MetaDescription, profile.MetaDescription)

	if err := s.db.Save(updates).Error; err != nil {
		return nil, err
	}
	return updates, nil
}
