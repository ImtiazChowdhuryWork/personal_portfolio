package models

import (
	"time"

	"gorm.io/gorm"
)

// JobPost stores a scraped job listing from any external source.
type JobPost struct {
	ID           uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	Source       string         `gorm:"type:varchar(50);not null" json:"source"` // remoteok | remotive | weworkremotely | linkedin
	Title        string         `gorm:"type:varchar(500);not null" json:"title"`
	Company      string         `gorm:"type:varchar(200)" json:"company"`
	Location     string         `gorm:"type:varchar(200)" json:"location"`
	JobType      string         `gorm:"type:varchar(100)" json:"job_type"`
	URL          string         `gorm:"type:varchar(1000);uniqueIndex" json:"url"`
	Description  string         `gorm:"type:text" json:"description"`
	PostedAt     time.Time      `json:"posted_at"`
	ScrapedAt    time.Time      `json:"scraped_at"`
	IsBookmarked bool           `gorm:"default:false" json:"is_bookmarked"`
	IsApplied    bool           `gorm:"default:false" json:"is_applied"`
	IsHidden     bool           `gorm:"default:false" json:"is_hidden"`

	// ─── Application Pipeline ─────────────────────────────────
	// Status: new | saved | applied | interviewing | offered | rejected | withdrawn
	Status      string     `gorm:"type:varchar(50);default:'new'" json:"status"`
	Notes       string     `gorm:"type:text" json:"notes"`
	AppliedAt   *time.Time `json:"applied_at"`
	InterviewAt *time.Time `json:"interview_at"`

	// ─── Recruiter Tracking ────────────────────────────────────
	RecruiterName     string `gorm:"type:varchar(200)" json:"recruiter_name"`
	RecruiterEmail    string `gorm:"type:varchar(200)" json:"recruiter_email"`
	RecruiterLinkedIn string `gorm:"type:varchar(500)" json:"recruiter_linkedin"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (JobPost) TableName() string {
	return "job_posts"
}
