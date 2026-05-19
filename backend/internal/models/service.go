// ============================================================
// FILE: internal/models/service.go
// WHAT IT IS:     Service/offering database model
// WHY IT EXISTS:  Stores the "What I Offer" cards shown on the
//                 portfolio's services section, managed from dashboard
// ENDPOINTS:      Used by service_handler.go for CRUD operations
// DEPENDS ON:     gorm.io/gorm
// LAST UPDATED:   2026-05-19 — initial creation
// ============================================================

package models

import (
	"time"

	"gorm.io/gorm"
)

// Service represents one "What I Offer" card on the portfolio.
type Service struct {
	ID          uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	Icon        string         `gorm:"type:varchar(100)" json:"icon"`
	Title       string         `gorm:"type:varchar(200);not null" json:"title"`
	Description string         `gorm:"type:text" json:"description"`
	SortOrder   int            `gorm:"default:0" json:"sort_order"`
	Active      bool           `gorm:"default:true" json:"active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
