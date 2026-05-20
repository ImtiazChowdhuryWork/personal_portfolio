package models

// ScraperSettings stores API credentials and scraper behaviour config.
// Always a single row (ID = 1).
type ScraperSettings struct {
	ID           uint   `gorm:"primaryKey" json:"id"`

	// ─── API Keys ─────────────────────────────────────────────
	AdzunaAppID  string `gorm:"type:varchar(200)" json:"adzuna_app_id"`
	AdzunaAppKey string `gorm:"type:varchar(200)" json:"adzuna_app_key"`
	RapidAPIKey  string `gorm:"type:varchar(500)" json:"rapidapi_key"`

	// ─── Scraper Behaviour ────────────────────────────────────
	ScrapeIntervalHours int    `gorm:"default:6"  json:"scrape_interval_hours"`
	MaxAgeDays          int    `gorm:"default:7"  json:"max_age_days"`
	DisabledSources     string `gorm:"type:text"  json:"disabled_sources"` // comma-separated source names

	// ─── Email Alerts ─────────────────────────────────────────
	AlertEnabled    bool `gorm:"default:false" json:"alert_enabled"`
	AlertMinNewJobs int  `gorm:"default:5"     json:"alert_min_new_jobs"`
}

func (ScraperSettings) TableName() string {
	return "scraper_settings"
}
