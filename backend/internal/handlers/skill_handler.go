// ============================================================
// FILE: internal/handlers/skill_handler.go
// WHAT IT IS:     HTTP handlers for the skills API
// ENDPOINTS:      GET    /api/v1/skills
//                 POST   /api/v1/skills    [protected]
//                 PUT    /api/v1/skills/:id [protected]
//                 DELETE /api/v1/skills/:id [protected]
// DEPENDS ON:     services/skill_service.go, utils/response.go
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

// SkillHandler holds the skill service dependency.
type SkillHandler struct {
	skillService *services.SkillService
}

// NewSkillHandler creates a new SkillHandler.
func NewSkillHandler(ss *services.SkillService) *SkillHandler {
	return &SkillHandler{skillService: ss}
}

// GetAll returns all skills grouped by category (public).
func (h *SkillHandler) GetAll(c *gin.Context) {
	skills, err := h.skillService.GetAll()
	if err != nil {
		utils.InternalError(c, "Failed to fetch skills")
		return
	}
	utils.Success(c, "Skills fetched successfully", skills)
}

// Create adds a new skill [protected].
func (h *SkillHandler) Create(c *gin.Context) {
	var skill models.Skill
	if err := c.ShouldBindJSON(&skill); err != nil {
		utils.BadRequest(c, "Invalid skill data", err.Error())
		return
	}
	if utils.IsEmpty(skill.Name) || utils.IsEmpty(skill.Category) {
		utils.BadRequest(c, "Skill name and category are required", nil)
		return
	}
	if !utils.IsValidPercentage(skill.Percentage) {
		utils.BadRequest(c, "Percentage must be between 0 and 100", nil)
		return
	}
	if err := h.skillService.Create(&skill); err != nil {
		utils.InternalError(c, "Failed to create skill")
		return
	}
	utils.Created(c, "Skill created successfully", skill)
}

// Update modifies an existing skill [protected].
func (h *SkillHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid skill ID", nil)
		return
	}
	var updates models.Skill
	if err := c.ShouldBindJSON(&updates); err != nil {
		utils.BadRequest(c, "Invalid skill data", err.Error())
		return
	}
	skill, err := h.skillService.Update(uint(id), &updates)
	if err != nil {
		utils.NotFound(c, err.Error())
		return
	}
	utils.Success(c, "Skill updated successfully", skill)
}

// Delete removes a skill [protected].
func (h *SkillHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid skill ID", nil)
		return
	}
	if err := h.skillService.Delete(uint(id)); err != nil {
		utils.NotFound(c, err.Error())
		return
	}
	utils.Success(c, "Skill deleted successfully", nil)
}
