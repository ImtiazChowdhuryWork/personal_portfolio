// ============================================================
// FILE: internal/models/message.go
// WHAT IT IS:     Contact form message database model
// WHY IT EXISTS:  Stores every message sent through the contact
//                 form so they can be reviewed in the dashboard
// ENDPOINTS:      Used by message_handler.go
// DEPENDS ON:     gorm.io/gorm
// IF REMOVED:     Contact form messages are lost — not stored anywhere
// LAST UPDATED:   2026-05-07 — initial creation
// ============================================================

package models

import (
	"time"

	"gorm.io/gorm"
)

// Message represents one contact form submission.
// Maps to the "messages" table in PostgreSQL.
type Message struct {
	// Primary key
	ID uint `gorm:"primaryKey;autoIncrement" json:"id"`

	// Name is what the visitor typed in the "Name" field
	Name string `gorm:"type:varchar(200);not null" json:"name"`

	// Email is the visitor's email address for replying
	Email string `gorm:"type:varchar(255);not null" json:"email"`

	// Type categorizes the inquiry:
	// "Job Opportunity" | "Freelance" | "Other"
	Type string `gorm:"type:varchar(100);default:'Other'" json:"type"`

	// Content is the full message body the visitor wrote
	Content string `gorm:"type:text;not null" json:"content"`

	// IsRead tracks whether Imtiaz has seen this message in the dashboard
	// Unread messages show a badge count in the dashboard sidebar
	IsRead bool `gorm:"default:false" json:"is_read"`

	// IsReplied tracks whether a reply was sent (for dashboard filtering)
	IsReplied bool `gorm:"default:false" json:"is_replied"`

	// IPAddress records where the message came from (for spam filtering)
	IPAddress string `gorm:"type:varchar(50)" json:"ip_address"`

	// GORM timestamp fields — CreatedAt is when the form was submitted
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
