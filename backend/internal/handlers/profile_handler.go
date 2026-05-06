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
