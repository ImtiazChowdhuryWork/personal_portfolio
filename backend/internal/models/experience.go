// ============================================================
// FILE: internal/models/experience.go
// WHAT IT IS:     Work experience database model
// WHY IT EXISTS:  Stores each job entry shown in the Experience
//                 Timeline section (SperkTech, SoftVence, DIU)
// ENDPOINTS:      Used by experience_handler.go
// DEPENDS ON:     gorm.io/gorm
// IF REMOVED:     Experience Timeline section has no data
// LAST UPDATED:   2026-05-07 — initial creation
// ============================================================

package models

import (
	"time"

	"gorm.io/gorm"
)

// Experience represents one job/internship on the timeline.
// Maps to the "experiences" table in PostgreSQL.
type Experience struct {
	// Primary key
	ID uint `gorm:"primaryKey;autoIncrement" json:"id"`

	// Company is the employer name (e.g. "SperkTech")
	Company string `gorm:"type:varchar(200);not null" json:"company"`

	// Role is the job title (e.g. "Flutter Developer")
	Role string `gorm:"type:varchar(200);not null" json:"role"`

	// StartDate is when the job began (stored as string "Jan 2022" for display)
	StartDate string `gorm:"type:varchar(50);not null" json:"start_date"`

	// EndDate is when the job ended — empty string means "Present"
	EndDate string `gorm:"type:varchar(50)" json:"end_date"`

	// IsCurrent marks the most recent job so the UI shows "Present" badge
	IsCurrent bool `gorm:"default:false" json:"is_current"`

	// Description is the main paragraph explaining the role
	Description string `gorm:"type:text" json:"description"`

	// Achievements stores the bullet points of specific accomplishments
	// Stored as JSON array: ["Built X", "Improved Y by Z%"]
	Achievements StringArray `gorm:"type:text" json:"achievements"`

	// TechUsed stores the technology tags shown on the timeline card
	// Example: ["Flutter", "Dart", "Firebase", "GetX"]
	TechUsed StringArray `gorm:"type:text" json:"tech_used"`

	// CompanyLogo is the path to the company's logo image
	CompanyLogo string `gorm:"type:varchar(500)" json:"company_logo"`

	// Location is where the job was (e.g. "Dhaka, Bangladesh")
	Location string `gorm:"type:varchar(200)" json:"location"`

	// Type is the employment type: "full-time" | "part-time" | "internship" | "freelance"
	Type string `gorm:"type:varchar(50);default:'full-time'" json:"type"`

	// SortOrder controls which experience shows first (lower = more recent, shown first)
	SortOrder int `gorm:"default:0" json:"sort_order"`

	// GORM timestamp fields
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
