package models

import "time"

// MailPasswordHistory records every Gmail App Password the admin has saved
// through the dashboard's Mail Settings page. The most recent row matches
// the value currently stored on the profile.
//
// SECURITY NOTE: stores the full App Password in plaintext at the admin's
// explicit request. Treat this table as sensitive — anyone with DB access
// can read every historical SMTP credential.
type MailPasswordHistory struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	GmailAddress string    `gorm:"type:varchar(255)" json:"gmail_address"`
	AppPassword  string    `gorm:"type:varchar(255);not null" json:"app_password"`
	CreatedAt    time.Time `json:"created_at"`

	// IsActive is computed at read time (gorm:"-" keeps it out of the table).
	// True only on the row whose password matches profile.smtp_pass right now.
	IsActive bool `gorm:"-" json:"is_active"`

	// IsHidden is set when this row's gmail_address appears in the
	// hidden_mail_accounts table. Used by the dashboard to dim the card.
	IsHidden bool `gorm:"-" json:"is_hidden"`
}
