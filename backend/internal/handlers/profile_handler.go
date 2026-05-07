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
	"imtiaz-portfolio/internal/models"
	"imtiaz-portfolio/internal/services"
	"imtiaz-portfolio/internal/utils"

	"github.com/gin-gonic/gin"
)

// ProfileHandler holds the profile service dependency.
type ProfileHandler struct {
	profileService *services.ProfileService
}

// NewProfileHandler creates a new ProfileHandler.
func NewProfileHandler(ps *services.ProfileService) *ProfileHandler {
	return &ProfileHandler{profileService: ps}
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
