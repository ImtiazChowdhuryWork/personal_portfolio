package models

import "time"

// HiddenMailAccount keeps a row per Gmail address the admin has chosen to
// hide from the dashboard's Available Accounts panel. The mail_password_histories
// rows for that email stay intact — only the deduped Available Accounts view
// filters them out so the audit log remains append-only.
type HiddenMailAccount struct {
	ID       uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Email    string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	HiddenAt time.Time `json:"hidden_at"`
}
