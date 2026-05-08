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

	"imtiaz-portfolio/internal/models"
	"imtiaz-portfolio/internal/services"
	"imtiaz-portfolio/internal/utils"

	"github.com/gin-gonic/gin"
)

// ProfileHandler holds the profile and email service dependencies.
type ProfileHandler struct {
	profileService *services.ProfileService
	emailService   *services.EmailService
}

// NewProfileHandler creates a new ProfileHandler.
func NewProfileHandler(ps *services.ProfileService, es *services.EmailService) *ProfileHandler {
	return &ProfileHandler{profileService: ps, emailService: es}
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

// UpdateCV updates only the cv_file field — protected, used by CV upload.
func (h *ProfileHandler) UpdateCV(c *gin.Context) {
	var body struct {
		CVFile string `json:"cv_file"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.CVFile == "" {
		utils.BadRequest(c, "cv_file is required", nil)
		return
	}
	if err := h.profileService.UpdateField("cv_file", body.CVFile); err != nil {
		utils.InternalError(c, "Failed to update CV")
		return
	}
	utils.Success(c, "CV updated successfully", gin.H{"cv_file": body.CVFile})
}
