package models

import "time"

// CVFile records every CV ever uploaded or auto-generated. Acts as a version
// history so the admin can switch between past CVs (Activate), preview them,
// or delete old versions. The currently active CV's path is mirrored on
// profile.cv_file so the public portfolio doesn't need a join.
type CVFile struct {
	ID        uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	FilePath  string `gorm:"type:varchar(500);not null"   json:"file_path"`
	FileName  string `gorm:"type:varchar(255);not null"   json:"file_name"`
	FileSize  int64  `gorm:"not null;default:0"           json:"file_size"`

	// Source is either "uploaded" (admin uploaded a PDF) or "generated"
	// (built from profile/skills/experience by the auto-generator).
	Source string `gorm:"type:varchar(20);not null;default:'uploaded'" json:"source"`

	CreatedAt time.Time `json:"created_at"`

	// IsActive is computed at read time (gorm:"-" keeps it out of the table).
	// True only on the row whose file_path matches profile.cv_file right now.
	IsActive bool `gorm:"-" json:"is_active"`
}
