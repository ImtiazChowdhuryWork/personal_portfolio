// ============================================================
// FILE: internal/handlers/profile_handler.go
// WHAT IT IS:     HTTP handlers for the portfolio profile API
// ENDPOINTS:      GET /api/v1/profile          (public)
//                 PUT /api/v1/profile          [protected]
// DEPENDS ON:     services/profile_service.go, utils/response.go
// LAST UPDATED:   2026-05-07 — initial creation
// ============================================================

package handlers

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"imtiaz-portfolio/internal/models"
	"imtiaz-portfolio/internal/services"
	"imtiaz-portfolio/internal/utils"

	"github.com/gin-gonic/gin"
)

// ProfileHandler holds the profile and email service dependencies.
// uploadDir is needed so DeleteCVHistory can remove the physical PDF file
// from disk when an old CV is purged from the version-history list.
type ProfileHandler struct {
	profileService *services.ProfileService
	emailService   *services.EmailService
	cvGenerator    *services.CVGeneratorService
	uploadDir      string
}

// NewProfileHandler creates a new ProfileHandler.
func NewProfileHandler(ps *services.ProfileService, es *services.EmailService, gen *services.CVGeneratorService, uploadDir string) *ProfileHandler {
	return &ProfileHandler{profileService: ps, emailService: es, cvGenerator: gen, uploadDir: uploadDir}
}

// Get returns the portfolio profile — public, called when the portfolio loads.
func (h *ProfileHandler) Get(c *gin.Context) {
	// Prevent caching so the latest profile_photo and cv_file always loads
	c.Header("Cache-Control", "no-store")
	profile, err := h.profileService.Get()
	if err != nil {
		utils.InternalError(c, "Failed to fetch profile")
		return
	}
	utils.Success(c, "Profile fetched successfully", profile)
}

// Update saves changes to the profile — protected, dashboard only.
func (h *ProfileHandler) Update(c *gin.Context) {
	var updates models.Profile
	if err := c.ShouldBindJSON(&updates); err != nil {
		utils.BadRequest(c, "Invalid profile data", err.Error())
		return
	}
	profile, err := h.profileService.Update(&updates)
	if err != nil {
		utils.InternalError(c, "Failed to update profile")
		return
	}
	utils.Success(c, "Profile updated successfully", profile)
}

// UpdatePhoto updates only the profile_photo field — protected, used by photo upload.
func (h *ProfileHandler) UpdatePhoto(c *gin.Context) {
	var body struct {
		ProfilePhoto string `json:"profile_photo"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.ProfilePhoto == "" {
		utils.BadRequest(c, "profile_photo is required", nil)
		return
	}
	if err := h.profileService.UpdateField("profile_photo", body.ProfilePhoto); err != nil {
		utils.InternalError(c, "Failed to update profile photo")
		return
	}
	utils.Success(c, "Profile photo updated successfully", gin.H{"profile_photo": body.ProfilePhoto})
}

// UpdateAboutPhoto updates only the about_photo field — protected.
func (h *ProfileHandler) UpdateAboutPhoto(c *gin.Context) {
	var body struct {
		AboutPhoto string `json:"about_photo"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.AboutPhoto == "" {
		utils.BadRequest(c, "about_photo is required", nil)
		return
	}
	if err := h.profileService.UpdateField("about_photo", body.AboutPhoto); err != nil {
		utils.InternalError(c, "Failed to update about photo")
		return
	}
	utils.Success(c, "About photo updated successfully", gin.H{"about_photo": body.AboutPhoto})
}

// UpdateSocial updates only the social-link fields — protected, used by the
// dashboard's Social Links section. Empty values are written through (allowing
// the admin to clear a link), which is intentionally different from the
// general-purpose Update handler that preserves untouched fields.
func (h *ProfileHandler) UpdateSocial(c *gin.Context) {
	var body struct {
		GitHub    *string `json:"github"`
		LinkedIn  *string `json:"linkedin"`
		Twitter   *string `json:"twitter"`
		Instagram *string `json:"instagram"`
		WhatsApp  *string `json:"whatsapp"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.BadRequest(c, "Invalid social link data", err.Error())
		return
	}
	fields := map[string]interface{}{}
	if body.GitHub != nil    { fields["git_hub"] = *body.GitHub }
	if body.LinkedIn != nil  { fields["linked_in"] = *body.LinkedIn }
	if body.Twitter != nil   { fields["twitter"] = *body.Twitter }
	if body.Instagram != nil { fields["instagram"] = *body.Instagram }
	if body.WhatsApp != nil  { fields["whats_app"] = *body.WhatsApp }
	if len(fields) == 0 {
		utils.BadRequest(c, "No social fields provided", nil)
		return
	}
	if err := h.profileService.UpdateFields(fields); err != nil {
		utils.InternalError(c, "Failed to update social links")
		return
	}
	profile, _ := h.profileService.Get()
	utils.Success(c, "Social links updated successfully", profile)
}

// UpdateGitHub updates only the GitHub Stats fields — protected, used by the
// dashboard's GitHub Stats section. Empty values are written through (which
// re-enables auto-mode for that card on the public portfolio).
func (h *ProfileHandler) UpdateGitHub(c *gin.Context) {
	var body struct {
		GitHubUsername    *string `json:"github_username"`
		GitHubRepos       *string `json:"github_repos"`
		GitHubCommits     *string `json:"github_commits"`
		GitHubTopLanguage *string `json:"github_top_language"`
		GitHubYearsActive *string `json:"github_years_active"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.BadRequest(c, "Invalid GitHub stats payload", err.Error())
		return
	}
	fields := map[string]interface{}{}
	if body.GitHubUsername != nil    { fields["git_hub_username"]     = *body.GitHubUsername }
	if body.GitHubRepos != nil       { fields["git_hub_repos"]        = *body.GitHubRepos }
	if body.GitHubCommits != nil     { fields["git_hub_commits"]      = *body.GitHubCommits }
	if body.GitHubTopLanguage != nil { fields["git_hub_top_language"] = *body.GitHubTopLanguage }
	if body.GitHubYearsActive != nil { fields["git_hub_years_active"] = *body.GitHubYearsActive }
	if len(fields) == 0 {
		utils.BadRequest(c, "No GitHub fields provided", nil)
		return
	}
	if err := h.profileService.UpdateFields(fields); err != nil {
		utils.InternalError(c, "Failed to update GitHub stats")
		return
	}
	profile, _ := h.profileService.Get()
	utils.Success(c, "GitHub stats updated successfully", profile)
}

// UpdateFooter updates only the four footer/copyright fields — protected,
// used by the dashboard's Footer tab. Empty strings are written through (so
// the admin can clear a value and re-enable the auto-fallback).
func (h *ProfileHandler) UpdateFooter(c *gin.Context) {
	var body struct {
		CopyrightText        *string `json:"copyright_text"`
		SidebarCopyrightText *string `json:"sidebar_copyright_text"`
		FooterCopyrightText  *string `json:"footer_copyright_text"`
		FooterBuiltWith      *string `json:"footer_built_with"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.BadRequest(c, "Invalid footer data", err.Error())
		return
	}
	fields := map[string]interface{}{}
	if body.CopyrightText        != nil { fields["copyright_text"]         = *body.CopyrightText }
	if body.SidebarCopyrightText != nil { fields["sidebar_copyright_text"] = *body.SidebarCopyrightText }
	if body.FooterCopyrightText  != nil { fields["footer_copyright_text"]  = *body.FooterCopyrightText }
	if body.FooterBuiltWith      != nil { fields["footer_built_with"]      = *body.FooterBuiltWith }
	if len(fields) == 0 {
		utils.BadRequest(c, "No footer fields provided", nil)
		return
	}
	if err := h.profileService.UpdateFields(fields); err != nil {
		utils.InternalError(c, "Failed to update footer")
		return
	}
	profile, _ := h.profileService.Get()
	utils.Success(c, "Footer updated successfully", profile)
}

// UpdateMail updates only the SMTP credentials — protected, used by the
// dashboard's Mail Settings section. Empty smtp_user is written through (so
// the admin can disable email replies). Empty smtp_pass is preserved so the
// admin doesn't accidentally wipe the password by saving with the boxes blank.
func (h *ProfileHandler) UpdateMail(c *gin.Context) {
	var body struct {
		SMTPUser *string `json:"smtp_user"`
		SMTPPass *string `json:"smtp_pass"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.BadRequest(c, "Invalid mail settings", err.Error())
		return
	}
	fields := map[string]interface{}{}
	if body.SMTPUser != nil {
		fields["smtp_user"] = *body.SMTPUser
	}
	if body.SMTPPass != nil && *body.SMTPPass != "" {
		// Strip any spaces Google may show in the App Password
		clean := *body.SMTPPass
		for i := 0; i < len(clean); i++ {
			if clean[i] == ' ' {
				clean = clean[:i] + clean[i+1:]
				i--
			}
		}
		fields["smtp_pass"] = clean
	}
	if len(fields) == 0 {
		utils.BadRequest(c, "No mail settings provided", nil)
		return
	}
	if err := h.profileService.UpdateFields(fields); err != nil {
		utils.InternalError(c, "Failed to update mail settings")
		return
	}
	// Record the password change in the history table when a new password was supplied.
	if newPass, ok := fields["smtp_pass"].(string); ok && newPass != "" {
		gmail := ""
		if v, ok := fields["smtp_user"].(string); ok {
			gmail = v
		} else {
			// User updated only the password — pull the current address for the history row
			if p, err := h.profileService.Get(); err == nil {
				gmail = p.SMTPUser
			}
		}
		_ = h.profileService.AddPasswordHistory(gmail, newPass)
	}
	profile, _ := h.profileService.Get()
	utils.Success(c, "Mail settings updated successfully", profile)
}

// GetMailHistory returns every saved Gmail App Password (newest first) — protected.
// The newest entry whose password matches the currently saved profile.smtp_pass
// is flagged as IsActive so the dashboard can render an "Active" badge. Each row
// is also flagged with IsHidden=true if its Gmail address appears in the
// hidden_mail_accounts table.
func (h *ProfileHandler) GetMailHistory(c *gin.Context) {
	history, err := h.profileService.GetPasswordHistory()
	if err != nil {
		utils.InternalError(c, "Failed to load password history")
		return
	}
	if profile, err := h.profileService.Get(); err == nil && profile.SMTPPass != "" {
		for i := range history {
			if history[i].AppPassword == profile.SMTPPass {
				history[i].IsActive = true
				break
			}
		}
	}
	if hidden, err := h.profileService.GetHiddenMailEmails(); err == nil {
		set := make(map[string]bool, len(hidden))
		for _, e := range hidden {
			set[e] = true
		}
		for i := range history {
			if set[history[i].GmailAddress] {
				history[i].IsHidden = true
			}
		}
	}
	utils.Success(c, "Password history fetched", history)
}

// VerifyMail checks supplied SMTP credentials against Gmail without sending mail.
// Always returns 200; the JSON body's `valid` flag tells the dashboard whether
// the password worked, and `error` carries the SMTP server's reason on failure.
func (h *ProfileHandler) VerifyMail(c *gin.Context) {
	var body struct {
		SMTPUser string `json:"smtp_user"`
		SMTPPass string `json:"smtp_pass"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.BadRequest(c, "Invalid request", err.Error())
		return
	}
	user := body.SMTPUser
	pass := body.SMTPPass
	// Strip any spaces Google may show in the App Password
	cleaned := make([]byte, 0, len(pass))
	for i := 0; i < len(pass); i++ {
		if pass[i] != ' ' {
			cleaned = append(cleaned, pass[i])
		}
	}
	pass = string(cleaned)
	if user == "" || pass == "" {
		utils.BadRequest(c, "Both smtp_user and smtp_pass are required", nil)
		return
	}
	if err := h.emailService.Verify(user, pass); err != nil {
		utils.Success(c, "Verification finished", gin.H{"valid": false, "error": err.Error()})
		return
	}
	utils.Success(c, "Verification finished", gin.H{"valid": true})
}

// resolveAccountEmail looks up the Gmail address that a given history-row id
// belongs to. Used by HideMailAccount and UnhideMailAccount so they can act on
// numeric ids (no path-encoding edge cases) instead of raw emails.
func (h *ProfileHandler) resolveAccountEmail(c *gin.Context) (string, bool) {
	idParam := c.Param("id")
	var id uint
	if _, err := fmt.Sscanf(idParam, "%d", &id); err != nil || id == 0 {
		utils.BadRequest(c, "Invalid history id", nil)
		return "", false
	}
	row, err := h.profileService.GetPasswordHistoryByID(id)
	if err != nil || row == nil || row.GmailAddress == "" {
		utils.BadRequest(c, "History row not found", nil)
		return "", false
	}
	return row.GmailAddress, true
}

// HideMailAccount marks an account as hidden from the Available Accounts
// panel. History rows for the email are preserved so the audit log stays intact.
// Refuses to hide the active sender to keep email replies working.
func (h *ProfileHandler) HideMailAccount(c *gin.Context) {
	email, ok := h.resolveAccountEmail(c)
	if !ok {
		return
	}
	log.Printf("[HideMailAccount] hiding email: %q", email)
	if profile, err := h.profileService.Get(); err == nil && profile.SMTPUser == email {
		utils.BadRequest(c,
			"This is the active sender. Switch to another account first, then hide it.",
			nil)
		return
	}
	if err := h.profileService.HideMailAccount(email); err != nil {
		log.Printf("[HideMailAccount] DB error for %q: %v", email, err)
		utils.InternalError(c, "Failed to hide account")
		return
	}
	utils.Success(c, "Account hidden", gin.H{"email": email})
}

// UnhideMailAccount restores a previously hidden account so it appears in
// Available Accounts again.
func (h *ProfileHandler) UnhideMailAccount(c *gin.Context) {
	email, ok := h.resolveAccountEmail(c)
	if !ok {
		return
	}
	log.Printf("[UnhideMailAccount] unhiding email: %q", email)
	if err := h.profileService.UnhideMailAccount(email); err != nil {
		log.Printf("[UnhideMailAccount] DB error for %q: %v", email, err)
		utils.InternalError(c, "Failed to unhide account")
		return
	}
	utils.Success(c, "Account restored", gin.H{"email": email})
}

// DeleteMailHistory removes one entry from the password history — protected.
func (h *ProfileHandler) DeleteMailHistory(c *gin.Context) {
	idParam := c.Param("id")
	var id uint
	if _, err := fmt.Sscanf(idParam, "%d", &id); err != nil || id == 0 {
		utils.BadRequest(c, "Invalid history id", nil)
		return
	}
	if err := h.profileService.DeletePasswordHistory(id); err != nil {
		utils.InternalError(c, "Failed to delete history entry")
		return
	}
	utils.Success(c, "History entry deleted", nil)
}

// UpdateCV updates only the cv_file field AND appends a history row to the
// cv_files table — protected, used by CV upload. Accepts optional file_name
// and file_size so the history list can show metadata.
func (h *ProfileHandler) UpdateCV(c *gin.Context) {
	var body struct {
		CVFile   string `json:"cv_file"`
		FileName string `json:"file_name"`
		FileSize int64  `json:"file_size"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.CVFile == "" {
		utils.BadRequest(c, "cv_file is required", nil)
		return
	}
	name := body.FileName
	if name == "" {
		// Fall back to the basename so older clients still produce a usable history row
		name = filepath.Base(body.CVFile)
	}
	row, err := h.profileService.AddCVHistory(body.CVFile, name, body.FileSize, "uploaded", true)
	if err != nil {
		utils.InternalError(c, "Failed to update CV")
		return
	}
	utils.Success(c, "CV updated successfully", gin.H{"cv_file": body.CVFile, "history": row})
}

// GetCVHistory returns every saved CV (newest first) — protected.
func (h *ProfileHandler) GetCVHistory(c *gin.Context) {
	history, err := h.profileService.GetCVHistory()
	if err != nil {
		utils.InternalError(c, "Failed to load CV history")
		return
	}
	utils.Success(c, "CV history fetched", history)
}

// ActivateCV makes a past CV the currently downloadable one — protected.
func (h *ProfileHandler) ActivateCV(c *gin.Context) {
	idParam := c.Param("id")
	var id uint
	if _, err := fmt.Sscanf(idParam, "%d", &id); err != nil || id == 0 {
		utils.BadRequest(c, "Invalid CV id", nil)
		return
	}
	row, err := h.profileService.ActivateCV(id)
	if err != nil {
		utils.InternalError(c, "Failed to activate CV")
		return
	}
	utils.Success(c, "CV activated", row)
}

// DeleteCVHistory removes one CV from the version history (and the physical
// file on disk) — protected. Refuses to delete the currently active CV so the
// public Download CV button never points at a missing file.
func (h *ProfileHandler) DeleteCVHistory(c *gin.Context) {
	idParam := c.Param("id")
	var id uint
	if _, err := fmt.Sscanf(idParam, "%d", &id); err != nil || id == 0 {
		utils.BadRequest(c, "Invalid CV id", nil)
		return
	}
	row, err := h.profileService.GetCVHistoryByID(id)
	if err != nil || row == nil {
		utils.BadRequest(c, "CV not found", nil)
		return
	}
	if profile, err := h.profileService.Get(); err == nil && profile.CVFile == row.FilePath {
		utils.BadRequest(c,
			"This is the active CV. Activate a different CV first, then delete this one.",
			nil)
		return
	}
	if err := h.profileService.DeleteCVHistory(id); err != nil {
		utils.InternalError(c, "Failed to delete CV")
		return
	}
	// Best-effort delete of the physical file. We strip the leading "/uploads/"
	// from the URL path and join it with the configured uploads directory.
	if strings.HasPrefix(row.FilePath, "/uploads/") {
		rel := strings.TrimPrefix(row.FilePath, "/uploads/")
		full, _ := filepath.Abs(filepath.Join(h.uploadDir, rel))
		_ = os.Remove(full)
	}
	utils.Success(c, "CV deleted", gin.H{"id": id})
}

// UpdateCVVisibility toggles whether the public Download CV button is shown
// — protected. Body: { "cv_visible": true | false }.
func (h *ProfileHandler) UpdateCVVisibility(c *gin.Context) {
	var body struct {
		CVVisible *bool `json:"cv_visible"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.CVVisible == nil {
		utils.BadRequest(c, "cv_visible (boolean) is required", nil)
		return
	}
	if err := h.profileService.UpdateField("cv_visible", *body.CVVisible); err != nil {
		utils.InternalError(c, "Failed to update CV visibility")
		return
	}
	utils.Success(c, "CV visibility updated", gin.H{"cv_visible": *body.CVVisible})
}

// GenerateCV builds a fresh PDF from the current profile + skills + experience
// and saves it as a new history row (active by default) — protected. Body is
// empty; returns the new history row.
func (h *ProfileHandler) GenerateCV(c *gin.Context) {
	if h.cvGenerator == nil {
		utils.InternalError(c, "CV generator not configured")
		return
	}
	row, err := h.cvGenerator.Generate()
	if err != nil {
		log.Printf("[GenerateCV] %v", err)
		utils.InternalError(c, "Failed to generate CV: "+err.Error())
		return
	}
	utils.Success(c, "CV generated", row)
}

// RecordCVDownload bumps the visitor download counter — PUBLIC. The public
// portfolio fires this just before navigating to the PDF. Returns the new
// count so the dashboard can poll without a separate read.
func (h *ProfileHandler) RecordCVDownload(c *gin.Context) {
	if err := h.profileService.IncrementCVDownloadCount(); err != nil {
		utils.InternalError(c, "Failed to record download")
		return
	}
	profile, _ := h.profileService.Get()
	count := int64(0)
	if profile != nil {
		count = profile.CVDownloadCount
	}
	utils.Success(c, "Download recorded", gin.H{"cv_download_count": count})
}
