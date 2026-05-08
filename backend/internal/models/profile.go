// ============================================================
// FILE: internal/models/profile.go
// WHAT IT IS:     Portfolio profile/bio database model
// WHY IT EXISTS:  Stores the editable profile data that appears
//                 across multiple sections of the portfolio
//                 (hero, about, contact, footer). Editing the
//                 profile in the dashboard updates all sections.
// ENDPOINTS:      Used by profile_handler.go
// DEPENDS ON:     gorm.io/gorm
// IF REMOVED:     Portfolio shows hardcoded text instead of dynamic content
// LAST UPDATED:   2026-05-07 — initial creation
// ============================================================

package models

import (
	"time"

	"gorm.io/gorm"
)

// Profile holds all the personalizable content of the portfolio.
// There is only ONE profile row — the admin's personal info.
// Maps to the "profiles" table in PostgreSQL.
type Profile struct {
	// Primary key — always 1 (single profile)
	ID uint `gorm:"primaryKey;autoIncrement" json:"id"`

	// ─── Basic Info ──────────────────────────────────────────
	FullName  string `gorm:"type:varchar(200);not null" json:"full_name"`
	Title     string `gorm:"type:varchar(200)" json:"title"`
	Tagline   string `gorm:"type:varchar(500)" json:"tagline"`
	Bio       string `gorm:"type:text" json:"bio"`
	ShortBio  string `gorm:"type:varchar(500)" json:"short_bio"`

	// ─── Contact Details ─────────────────────────────────────
	Email     string `gorm:"type:varchar(255)" json:"email"`
	Phone     string `gorm:"type:varchar(50)" json:"phone"`
	WhatsApp  string `gorm:"type:varchar(50)" json:"whatsapp"`
	Location  string `gorm:"type:varchar(200)" json:"location"`

	// ─── Social Links ────────────────────────────────────────
	GitHub    string `gorm:"type:varchar(500)" json:"github"`
	LinkedIn  string `gorm:"type:varchar(500)" json:"linkedin"`
	Twitter   string `gorm:"type:varchar(500)" json:"twitter"`
	Instagram string `gorm:"type:varchar(500)" json:"instagram"`

	// ─── Media ───────────────────────────────────────────────
	// ProfilePhoto is shown in the sidebar card
	ProfilePhoto string `gorm:"type:varchar(500)" json:"profile_photo"`

	// AboutPhoto is shown in the About section (can be a different image)
	AboutPhoto string `gorm:"type:varchar(500)" json:"about_photo"`

	// CVFile is the path to the downloadable CV PDF
	CVFile string `gorm:"type:varchar(500)" json:"cv_file"`

	// ─── Status ──────────────────────────────────────────────
	// Availability is shown in the About section info card
	// Example: "Open to Work", "Busy", "Available for Freelance"
	Availability string `gorm:"type:varchar(100);default:'Open to Work'" json:"availability"`

	// YearsExperience is shown in the stat counters
	YearsExperience string `gorm:"type:varchar(20);default:'2.5+'" json:"years_experience"`

	// AppsShipped is shown in the stat counters
	AppsShipped string `gorm:"type:varchar(20);default:'5+'" json:"apps_shipped"`

	// ─── Reply From Emails ───────────────────────────────────
	// Comma-separated list of email addresses shown in the dashboard
	// reply compose dropdown so the admin can pick which address to send from.
	// Example: "work@gmail.com, personal@gmail.com, imtiaz@company.com"
	ReplyEmails string `gorm:"type:text" json:"reply_emails"`

	// ─── SMTP / Mail Settings ─────────────────────────────────
	// Stored in DB so the admin can update them from the dashboard
	// without touching the .env file. Overrides .env if non-empty.
	SMTPUser string `gorm:"type:varchar(255)" json:"smtp_user"`
	SMTPPass string `gorm:"type:varchar(255)" json:"smtp_pass"`

	// ─── GitHub Stats ─────────────────────────────────────────
	// GitHubUsername drives the live API fetch in the public portfolio.
	// The four override fields are blank by default — when blank, the
	// frontend uses the auto-computed value from GitHub. Setting any of
	// them substitutes that value in the matching stat card.
	GitHubUsername    string `gorm:"type:varchar(100)" json:"github_username"`
	GitHubRepos       string `gorm:"type:varchar(20)"  json:"github_repos"`
	GitHubCommits     string `gorm:"type:varchar(20)"  json:"github_commits"`
	GitHubTopLanguage string `gorm:"type:varchar(50)"  json:"github_top_language"`
	GitHubYearsActive string `gorm:"type:varchar(20)"  json:"github_years_active"`

	// ─── SEO ─────────────────────────────────────────────────
	MetaTitle       string `gorm:"type:varchar(200)" json:"meta_title"`
	MetaDescription string `gorm:"type:varchar(500)" json:"meta_description"`

	// GORM timestamp fields
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
