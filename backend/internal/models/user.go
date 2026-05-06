// ============================================================
// FILE: internal/models/user.go
// WHAT IT IS:     User database model (the admin account)
// WHY IT EXISTS:  Defines the structure of the "users" table.
//                 The portfolio has one admin user who logs into
//                 the dashboard to manage content.
// ENDPOINTS:      Used by auth_handler.go for login/logout
// DEPENDS ON:     gorm.io/gorm
// IF REMOVED:     Authentication breaks — no user to log in as
// LAST UPDATED:   2026-05-07 — initial creation
// ============================================================

package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents the admin account stored in the "users" table.
// GORM reads this struct to know which columns exist in the database.
// The `json:"-"` tags hide sensitive fields from API responses.
type User struct {
	// ID is the primary key — GORM auto-increments this
	ID uint `gorm:"primaryKey;autoIncrement" json:"id"`

	// Name shown in the dashboard header
	Name string `gorm:"type:varchar(100);not null" json:"name"`

	// Email used to log in — must be unique across all users
	Email string `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`

	// Password is stored as a bcrypt hash — NEVER stored in plain text
	// json:"-" means this field is NEVER included in JSON responses
	Password string `gorm:"type:varchar(255);not null" json:"-"`

	// Role controls what the user can do — currently always "admin"
	Role string `gorm:"type:varchar(50);default:'admin'" json:"role"`

	// Avatar is the path to the user's profile picture file
	Avatar string `gorm:"type:varchar(500)" json:"avatar"`

	// GORM automatically manages these three timestamp fields:
	// CreatedAt — set when the row is first inserted
	// UpdatedAt — updated every time the row changes
	// DeletedAt — set when soft-deleted (row stays in DB but is hidden)
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
