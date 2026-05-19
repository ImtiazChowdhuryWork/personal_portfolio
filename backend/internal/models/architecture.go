// ============================================================
// FILE: internal/models/architecture.go
// WHAT IT IS:     Architecture expertise database model
// WHY IT EXISTS:  Stores the "Architecture" cards shown on the
//                 portfolio, managed from the dashboard
// LAST UPDATED:   2026-05-19 — initial creation
// ============================================================

package models

import (
	"time"

	"gorm.io/gorm"
)

// Architecture represents one architecture pattern card on the portfolio.
type Architecture struct {
	ID          uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	Number      string         `gorm:"type:varchar(10)" json:"number"`
	Name        string         `gorm:"type:varchar(200);not null" json:"name"`
	Diagram     string         `gorm:"type:text" json:"diagram"`
	Description string         `gorm:"type:text" json:"description"`
	Projects    StringArray    `gorm:"type:text" json:"projects"`
	SortOrder   int            `gorm:"default:0" json:"sort_order"`
	Active      bool           `gorm:"default:true" json:"active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
