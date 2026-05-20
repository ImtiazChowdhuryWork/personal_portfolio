package models

import "time"

// JobScrapeLog records the outcome of one scrape run for a single source.
type JobScrapeLog struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Source     string    `gorm:"type:varchar(50);not null;index" json:"source"`
	ScrapedAt  time.Time `json:"scraped_at"`
	JobsFound  int       `json:"jobs_found"`
	JobsAdded  int       `json:"jobs_added"`
	ErrorMsg   string    `gorm:"type:text" json:"error_msg"`
	DurationMs int64     `json:"duration_ms"`
}

func (JobScrapeLog) TableName() string {
	return "job_scrape_logs"
}
