package services

import (
	"errors"
	"imtiaz-portfolio/internal/models"
	"strings"
	"time"

	"gorm.io/gorm"
)

// JobService handles all database operations for scraped job posts.
type JobService struct {
	db *gorm.DB
}

// NewJobService creates a new JobService with a database connection.
func NewJobService(db *gorm.DB) *JobService {
	return &JobService{db: db}
}

// GetAll fetches non-hidden jobs with optional filters:
// source — exact match on source name
// q      — keyword search across title, company, location
// bookmarked / applied — boolean flags
// getMaxAgeDays returns the configured max job age, defaulting to 7.
func (s *JobService) getMaxAgeDays() int {
	settings, err := s.GetScraperSettings()
	if err != nil || settings == nil || settings.MaxAgeDays <= 0 {
		return 7
	}
	return settings.MaxAgeDays
}

func (s *JobService) GetAll(source, q string, bookmarked, applied bool) ([]models.JobPost, error) {
	cutoff := time.Now().AddDate(0, 0, -s.getMaxAgeDays())
	query := s.db.Where("is_hidden = ? AND posted_at >= ?", false, cutoff)

	if source != "" {
		query = query.Where("source = ?", source)
	}
	if q != "" {
		like := "%" + strings.ToLower(q) + "%"
		query = query.Where(
			"LOWER(title) LIKE ? OR LOWER(company) LIKE ? OR LOWER(location) LIKE ?",
			like, like, like,
		)
	}
	if bookmarked {
		query = query.Where("is_bookmarked = ?", true)
	}
	if applied {
		query = query.Where("is_applied = ?", true)
	}

	var jobs []models.JobPost
	result := query.Order("posted_at DESC, scraped_at DESC").Find(&jobs)
	return jobs, result.Error
}

// ToggleBookmark flips the is_bookmarked flag on a job.
func (s *JobService) ToggleBookmark(id uint) error {
	result := s.db.Model(&models.JobPost{}).
		Where("id = ? AND is_hidden = ?", id, false).
		Update("is_bookmarked", gorm.Expr("NOT is_bookmarked"))
	if result.RowsAffected == 0 {
		return errors.New("job not found")
	}
	return result.Error
}

// ToggleApplied flips the is_applied flag on a job.
func (s *JobService) ToggleApplied(id uint) error {
	result := s.db.Model(&models.JobPost{}).
		Where("id = ? AND is_hidden = ?", id, false).
		Update("is_applied", gorm.Expr("NOT is_applied"))
	if result.RowsAffected == 0 {
		return errors.New("job not found")
	}
	return result.Error
}

// DeleteJob soft-hides a job (sets is_hidden = true so it disappears from lists).
func (s *JobService) DeleteJob(id uint) error {
	result := s.db.Model(&models.JobPost{}).
		Where("id = ?", id).
		Update("is_hidden", true)
	if result.RowsAffected == 0 {
		return errors.New("job not found")
	}
	return result.Error
}

// GetKeywords returns all stored priority keywords sorted alphabetically.
func (s *JobService) GetKeywords() ([]string, error) {
	var rows []models.JobKeyword
	if err := s.db.Order("keyword ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]string, len(rows))
	for i, r := range rows {
		out[i] = r.Keyword
	}
	return out, nil
}

// SetKeywords replaces the full keyword list atomically.
// Words are trimmed and lower-cased; blank entries are ignored.
func (s *JobService) SetKeywords(words []string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id > ?", 0).Delete(&models.JobKeyword{}).Error; err != nil {
			return err
		}
		seen := make(map[string]bool)
		for _, w := range words {
			w = strings.TrimSpace(strings.ToLower(w))
			if w == "" || seen[w] {
				continue
			}
			seen[w] = true
			if err := tx.Create(&models.JobKeyword{Keyword: w}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// SaveJobs inserts a batch of scraped jobs, skipping any URL that already exists.
// Returns the count of newly inserted rows.
func (s *JobService) SaveJobs(jobs []models.JobPost) (int, error) {
	added := 0
	for i := range jobs {
		var count int64
		s.db.Model(&models.JobPost{}).Unscoped().Where("url = ?", jobs[i].URL).Count(&count)
		if count > 0 {
			continue
		}
		if err := s.db.Create(&jobs[i]).Error; err == nil {
			added++
		}
	}
	return added, nil
}

// GetScraperSettings returns the settings row, or nil if never saved yet.
func (s *JobService) GetScraperSettings() (*models.ScraperSettings, error) {
	var row models.ScraperSettings
	err := s.db.First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil // not configured yet — caller treats nil as "use defaults"
	}
	return &row, err
}

// SaveScraperSettingsRequest is the shape of a settings update from the dashboard.
type SaveScraperSettingsRequest struct {
	// API keys
	AdzunaAppID  string `json:"adzuna_app_id"`
	AdzunaAppKey string `json:"adzuna_app_key"`
	RapidAPIKey  string `json:"rapidapi_key"`
	ClearAdzuna  bool   `json:"clear_adzuna"`
	ClearJSearch bool   `json:"clear_jsearch"`
	// Scraper behaviour
	ScrapeIntervalHours *int     `json:"scrape_interval_hours"` // pointer so 0 is valid
	MaxAgeDays          *int     `json:"max_age_days"`
	DisabledSources     *string  `json:"disabled_sources"`      // comma-separated; nil = don't change
	// Email alert
	AlertEnabled    *bool `json:"alert_enabled"`
	AlertMinNewJobs *int  `json:"alert_min_new_jobs"`
}

// SaveScraperSettings upserts the single settings row (always ID = 1).
// Empty string = keep existing value; ClearX = true explicitly wipes that source's keys.
func (s *JobService) SaveScraperSettings(req SaveScraperSettingsRequest) error {
	// Load existing row so we only overwrite fields the user touched
	var row models.ScraperSettings
	if err := s.db.First(&row).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		row.ID = 1 // first save ever — start from blank
	}

	if req.ClearAdzuna {
		row.AdzunaAppID = ""
		row.AdzunaAppKey = ""
	} else {
		if req.AdzunaAppID != "" {
			row.AdzunaAppID = req.AdzunaAppID
		}
		if req.AdzunaAppKey != "" {
			row.AdzunaAppKey = req.AdzunaAppKey
		}
	}

	if req.ClearJSearch {
		row.RapidAPIKey = ""
	} else if req.RapidAPIKey != "" {
		row.RapidAPIKey = req.RapidAPIKey
	}

	if req.ScrapeIntervalHours != nil {
		row.ScrapeIntervalHours = *req.ScrapeIntervalHours
	}
	if req.MaxAgeDays != nil {
		row.MaxAgeDays = *req.MaxAgeDays
	}
	if req.DisabledSources != nil {
		row.DisabledSources = *req.DisabledSources
	}
	if req.AlertEnabled != nil {
		row.AlertEnabled = *req.AlertEnabled
	}
	if req.AlertMinNewJobs != nil {
		row.AlertMinNewJobs = *req.AlertMinNewJobs
	}

	return s.db.Save(&row).Error
}

// UpdateStatus sets the application pipeline status and optional dates.
func (s *JobService) UpdateStatus(id uint, status string, appliedAt, interviewAt *time.Time) error {
	updates := map[string]interface{}{"status": status}
	if appliedAt != nil {
		updates["applied_at"] = appliedAt
	}
	if interviewAt != nil {
		updates["interview_at"] = interviewAt
	}
	result := s.db.Model(&models.JobPost{}).Where("id = ?", id).Updates(updates)
	if result.RowsAffected == 0 {
		return errors.New("job not found")
	}
	return result.Error
}

// UpdateNotes saves free-text notes on a job.
func (s *JobService) UpdateNotes(id uint, notes string) error {
	result := s.db.Model(&models.JobPost{}).Where("id = ?", id).
		Update("notes", notes)
	if result.RowsAffected == 0 {
		return errors.New("job not found")
	}
	return result.Error
}

// UpdateRecruiter saves recruiter contact info on a job.
func (s *JobService) UpdateRecruiter(id uint, name, email, linkedin string) error {
	result := s.db.Model(&models.JobPost{}).Where("id = ?", id).Updates(map[string]interface{}{
		"recruiter_name":      name,
		"recruiter_email":     email,
		"recruiter_linkedin":  linkedin,
	})
	if result.RowsAffected == 0 {
		return errors.New("job not found")
	}
	return result.Error
}

// JobStats holds aggregated counts for the analytics bar.
type JobStats struct {
	TotalJobs    int64 `json:"total_jobs"`
	NewToday     int64 `json:"new_today"`
	NewThisWeek  int64 `json:"new_this_week"`
	Saved        int64 `json:"saved"`
	Applied      int64 `json:"applied"`
	Interviewing int64 `json:"interviewing"`
	Offered      int64 `json:"offered"`
	Rejected     int64 `json:"rejected"`
}

// GetStats returns aggregated job counts for the analytics bar.
// Each count uses a fresh query — sharing a base *gorm.DB and chaining Where()
// mutates the statement, causing each count to accumulate all previous conditions.
func (s *JobService) GetStats() (*JobStats, error) {
	stats := &JobStats{}
	q := func() *gorm.DB { return s.db.Model(&models.JobPost{}).Where("is_hidden = ?", false) }

	todayStart := time.Now().Truncate(24 * time.Hour)
	weekStart  := time.Now().AddDate(0, 0, -7)

	q().Count(&stats.TotalJobs)
	q().Where("scraped_at >= ?", todayStart).Count(&stats.NewToday)
	q().Where("scraped_at >= ?", weekStart).Count(&stats.NewThisWeek)
	q().Where("is_bookmarked = ?", true).Count(&stats.Saved)
	q().Where("status = ?", "applied").Count(&stats.Applied)
	q().Where("status = ?", "interviewing").Count(&stats.Interviewing)
	q().Where("status = ?", "offered").Count(&stats.Offered)
	q().Where("status = ?", "rejected").Count(&stats.Rejected)

	return stats, nil
}

// GetLatestScrapeLogs returns the most recent log entry per source.
func (s *JobService) GetLatestScrapeLogs() ([]models.JobScrapeLog, error) {
	// Subquery: max ID per source group identifies the latest row
	sub := s.db.Model(&models.JobScrapeLog{}).Select("MAX(id)").Group("source")
	var logs []models.JobScrapeLog
	err := s.db.Where("id IN (?)", sub).Order("source ASC").Find(&logs).Error
	return logs, err
}
