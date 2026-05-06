// ============================================================
// FILE: internal/handlers/experience_handler.go
// WHAT IT IS:     HTTP handlers for the experience timeline API
// ENDPOINTS:      GET    /api/v1/experience
//                 POST   /api/v1/experience    [protected]
//                 PUT    /api/v1/experience/:id [protected]
//                 DELETE /api/v1/experience/:id [protected]
// DEPENDS ON:     services/experience_service.go, utils/response.go
// LAST UPDATED:   2026-05-07 — initial creation
// ============================================================

package handlers

import (
	"imtiaz-portfolio/internal/models"
	"imtiaz-portfolio/internal/services"
	"imtiaz-portfolio/internal/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ExperienceHandler holds the experience service dependency.
type ExperienceHandler struct {
	expService *services.ExperienceService
}

// NewExperienceHandler creates a new ExperienceHandler.
func NewExperienceHandler(es *services.ExperienceService) *ExperienceHandler {
	return &ExperienceHandler{expService: es}
}

// GetAll returns all experience entries (public).
func (h *ExperienceHandler) GetAll(c *gin.Context) {
	experiences, err := h.expService.GetAll()
	if err != nil {
		utils.InternalError(c, "Failed to fetch experience")
		return
	}
	utils.Success(c, "Experience fetched successfully", experiences)
}

// Create adds a new experience entry [protected].
func (h *ExperienceHandler) Create(c *gin.Context) {
	var exp models.Experience
	if err := c.ShouldBindJSON(&exp); err != nil {
		utils.BadRequest(c, "Invalid experience data", err.Error())
		return
	}
	if utils.IsEmpty(exp.Company) || utils.IsEmpty(exp.Role) {
		utils.BadRequest(c, "Company name and role are required", nil)
		return
	}
	if err := h.expService.Create(&exp); err != nil {
		utils.InternalError(c, "Failed to create experience entry")
		return
	}
	utils.Created(c, "Experience created successfully", exp)
}

// Update modifies an existing experience entry [protected].
func (h *ExperienceHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid experience ID", nil)
		return
	}
	var updates models.Experience
	if err := c.ShouldBindJSON(&updates); err != nil {
		utils.BadRequest(c, "Invalid experience data", err.Error())
		return
	}
	exp, err := h.expService.Update(uint(id), &updates)
	if err != nil {
		utils.NotFound(c, err.Error())
		return
	}
	utils.Success(c, "Experience updated successfully", exp)
}

// Delete removes an experience entry [protected].
func (h *ExperienceHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid experience ID", nil)
		return
	}
	if err := h.expService.Delete(uint(id)); err != nil {
		utils.NotFound(c, err.Error())
		return
	}
	utils.Success(c, "Experience deleted successfully", nil)
}
