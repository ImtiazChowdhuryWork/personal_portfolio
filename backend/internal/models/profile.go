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

	// Nickname is the short greeting name shown in the hero subtitle pill
	// ("Say Hi from {nickname}, {title}"). Empty = first word of FullName.
	Nickname string `gorm:"type:varchar(100)" json:"nickname"`

	// HeroSubtitle is a full override of the hero pill text. When set, it
	// replaces the default "Say Hi from {nickname}, {title}" template.
	// Supports placeholders {name}, {title}, {nickname}.
	HeroSubtitle string `gorm:"type:varchar(300)" json:"hero_subtitle"`

	// ─── Hero Heading ────────────────────────────────────────
	// Three plain lines, plus a comma-separated list of words to highlight in
	// the brand colour wherever they appear. HeroHeadingOverride takes
	// precedence when set: each newline is rendered as a line break and any
	// text wrapped in *asterisks* becomes a highlight span.
	HeroHeadingLine1      string `gorm:"type:varchar(200)" json:"hero_heading_line1"`
	HeroHeadingLine2      string `gorm:"type:varchar(200)" json:"hero_heading_line2"`
	HeroHeadingLine3      string `gorm:"type:varchar(200)" json:"hero_heading_line3"`
	HeroHeadingHighlights string `gorm:"type:varchar(500)" json:"hero_heading_highlights"`
	HeroHeadingOverride   string `gorm:"type:text"         json:"hero_heading_override"`

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

	// CVFile is the path to the currently active CV PDF (mirrored from cv_files).
	CVFile string `gorm:"type:varchar(500)" json:"cv_file"`

	// CVVisible controls whether the "Download CV" button is shown on the
	// public portfolio. Lets the admin temporarily hide the button without
	// deleting the active CV.
	CVVisible bool `gorm:"default:true" json:"cv_visible"`

	// CVDownloadCount is incremented every time a visitor clicks the
	// "Download CV" button on the public portfolio.
	CVDownloadCount int64 `gorm:"default:0" json:"cv_download_count"`

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

	// ─── Footer ──────────────────────────────────────────────
	// CopyrightText is the default copyright shown in BOTH the sidebar and the
	// footer, unless one of the per-place overrides below is set. Empty means
	// auto-fallback ("© {year} {full_name}. All Rights Reserved.").
	// Supports placeholders {year} and {name} which the frontend replaces.
	CopyrightText string `gorm:"type:varchar(200)" json:"copyright_text"`

	// SidebarCopyrightText overrides CopyrightText for the sidebar only.
	SidebarCopyrightText string `gorm:"type:varchar(200)" json:"sidebar_copyright_text"`

	// FooterCopyrightText overrides CopyrightText for the footer only.
	FooterCopyrightText string `gorm:"type:varchar(200)" json:"footer_copyright_text"`

	// FooterBuiltWith is a small second line shown in the footer below the
	// copyright, e.g. "Built with Flutter spirit 💙". Empty hides the line.
	FooterBuiltWith string `gorm:"type:varchar(200)" json:"footer_built_with"`

	// ─── SEO ─────────────────────────────────────────────────
	MetaTitle       string `gorm:"type:varchar(200)" json:"meta_title"`
	MetaDescription string `gorm:"type:varchar(500)" json:"meta_description"`

	// GORM timestamp fields
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
