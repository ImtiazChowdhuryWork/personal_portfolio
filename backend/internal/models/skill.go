// ============================================================
// FILE: internal/models/skill.go
// WHAT IT IS:     Skill database model for the Tech Stack section
// WHY IT EXISTS:  Stores each technology (Flutter, Dart, Firebase
//                 etc.) with its proficiency percentage so the
//                 frontend can render animated skill bars
// ENDPOINTS:      Used by skill_handler.go
// DEPENDS ON:     gorm.io/gorm
// IF REMOVED:     Tech Stack section shows no skills
// LAST UPDATED:   2026-05-07 — initial creation
// ============================================================

package models

import (
	"time"

	"gorm.io/gorm"
)

// Skill represents one technology entry in the Tech Stack section.
// Maps to the "skills" table in PostgreSQL.
type Skill struct {
	// Primary key — auto-incremented
	ID uint `gorm:"primaryKey;autoIncrement" json:"id"`

	// The technology name shown to visitors (e.g. "Flutter")
	Name string `gorm:"type:varchar(100);not null" json:"name"`

	// Category groups skills together in the UI:
	// "core" | "state_management" | "backend" | "payments" | "tools"
	Category string `gorm:"type:varchar(100);not null" json:"category"`

	// Percentage is the skill level shown on the bar (0-100)
	// Example: Flutter = 95, WebRTC = 75
	Percentage int `gorm:"not null;check:percentage >= 0 AND percentage <= 100" json:"percentage"`

	// Description appears on hover explaining the technology briefly
	// Example: "Cross-platform UI framework by Google"
	Description string `gorm:"type:text" json:"description"`

	// Icon is the path or URL to the technology's logo image
	// Example: "/assets/images/logos/flutter.svg"
	Icon string `gorm:"type:varchar(500)" json:"icon"`

	// SortOrder controls display order within each category
	SortOrder int `gorm:"default:0" json:"sort_order"`

	// GORM timestamp fields
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
