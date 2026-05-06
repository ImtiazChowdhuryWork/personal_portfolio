// ============================================================
// FILE: internal/models/project.go
// WHAT IT IS:     Project/App database model
// WHY IT EXISTS:  Stores all portfolio apps (Project Finder,
//                 Hiye Health, etc.) that are displayed in the
//                 Apps Showcase section of the portfolio
// ENDPOINTS:      Used by project_handler.go for CRUD operations
// DEPENDS ON:     gorm.io/gorm, database/sql, encoding/json
// IF REMOVED:     Apps Showcase section has no data to display
// LAST UPDATED:   2026-05-07 — initial creation
// ============================================================

package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// StringArray is a custom type that stores a Go string slice as JSON
// in a single PostgreSQL text column.
// Example: ["Flutter", "Firebase", "BLoC"] stored as a JSON string in the DB
type StringArray []string

/**
 * FUNCTION: Value (StringArray)
 * WHAT IT DOES:   Converts a StringArray into a JSON string so
 *                 PostgreSQL can store it as text
 * WHERE CALLED:   GORM calls this automatically before every INSERT/UPDATE
 * LAST UPDATED:   2026-05-07 — initial creation
 */
func (s StringArray) Value() (driver.Value, error) {
	if s == nil {
		return "[]", nil
	}
	bytes, err := json.Marshal(s)
	return string(bytes), err
}

/**
 * FUNCTION: Scan (StringArray)
 * WHAT IT DOES:   Converts a JSON string from PostgreSQL back into
 *                 a Go string slice when reading from the database
 * WHERE CALLED:   GORM calls this automatically after every SELECT
 * LAST UPDATED:   2026-05-07 — initial creation
 */
func (s *StringArray) Scan(value interface{}) error {
	if value == nil {
		*s = StringArray{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		// Try string type as well — PostgreSQL drivers vary
		str, ok := value.(string)
		if !ok {
			return errors.New("type assertion to []byte or string failed")
		}
		bytes = []byte(str)
	}
	return json.Unmarshal(bytes, s)
}

// Project represents one app/project entry in the portfolio.
// Maps directly to the "projects" table in PostgreSQL.
type Project struct {
	// Primary key — auto-incremented by PostgreSQL
	ID uint `gorm:"primaryKey;autoIncrement" json:"id"`

	// The app's display name (e.g. "Project Finder")
	Name string `gorm:"type:varchar(200);not null" json:"name"`

	// Slug is a URL-friendly version of the name used in API routes
	// Example: "project-finder"
	Slug string `gorm:"type:varchar(200);uniqueIndex" json:"slug"`

	// Short description shown in the app card (2-3 sentences)
	Description string `gorm:"type:text" json:"description"`

	// Long description shown in the detail modal
	LongDescription string `gorm:"type:text" json:"long_description"`

	// TechStack stores ["Flutter", "Firebase", "BLoC"] as JSON text
	TechStack StringArray `gorm:"type:text" json:"tech_stack"`

	// Features stores bullet points ["Real-time updates", ...] as JSON
	Features StringArray `gorm:"type:text" json:"features"`

	// Screenshots stores file paths or URLs of app screenshots
	Screenshots StringArray `gorm:"type:text" json:"screenshots"`

	// Thumbnail is the main preview image path for the app card
	Thumbnail string `gorm:"type:varchar(500)" json:"thumbnail"`

	// AppStoreURL links to the iOS App Store listing (can be empty)
	AppStoreURL string `gorm:"type:varchar(500)" json:"app_store_url"`

	// PlayStoreURL links to the Google Play Store listing (can be empty)
	PlayStoreURL string `gorm:"type:varchar(500)" json:"play_store_url"`

	// GithubURL links to the source code repository (can be empty if private)
	GithubURL string `gorm:"type:varchar(500)" json:"github_url"`

	// Status tells visitors if this app is live, in development, etc.
	// Values: "live", "development", "archived"
	Status string `gorm:"type:varchar(50);default:'live'" json:"status"`

	// Featured marks apps that should appear first in the showcase slider
	Featured bool `gorm:"default:false" json:"featured"`

	// SortOrder controls which app appears first (lower = first)
	SortOrder int `gorm:"default:0" json:"sort_order"`

	// GORM timestamp fields — managed automatically
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
