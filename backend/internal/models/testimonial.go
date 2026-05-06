// ============================================================
// FILE: internal/models/testimonial.go
// WHAT IT IS:     Testimonial database model
// WHY IT EXISTS:  Stores reviews and endorsements from clients
//                 or colleagues that add social proof to the portfolio
// ENDPOINTS:      Used by testimonial_handler.go
// DEPENDS ON:     gorm.io/gorm
// IF REMOVED:     Testimonials section has no data (section can be hidden)
// LAST UPDATED:   2026-05-07 — initial creation
// ============================================================

package models

import (
	"time"

	"gorm.io/gorm"
)

// Testimonial represents one review or endorsement.
// Maps to the "testimonials" table in PostgreSQL.
type Testimonial struct {
	// Primary key
	ID uint `gorm:"primaryKey;autoIncrement" json:"id"`

	// AuthorName is the full name of the person giving the review
	AuthorName string `gorm:"type:varchar(200);not null" json:"author_name"`

	// AuthorTitle is their job title (e.g. "Product Manager at SperkTech")
	AuthorTitle string `gorm:"type:varchar(200)" json:"author_title"`

	// AuthorAvatar is the path to their profile photo
	AuthorAvatar string `gorm:"type:varchar(500)" json:"author_avatar"`

	// Content is the actual review text
	Content string `gorm:"type:text;not null" json:"content"`

	// Rating is 1-5 stars (used to render star icons in the UI)
	Rating int `gorm:"default:5;check:rating >= 1 AND rating <= 5" json:"rating"`

	// IsPublished controls whether this testimonial shows on the portfolio
	// Unpublished ones are visible only in the admin dashboard
	IsPublished bool `gorm:"default:false" json:"is_published"`

	// SortOrder controls display order (lower = shown first)
	SortOrder int `gorm:"default:0" json:"sort_order"`

	// GORM timestamp fields
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
