// ============================================================
// FILE: internal/handlers/stats_handler.go
// WHAT IT IS:     HTTP handler for dashboard statistics endpoint
// ENDPOINTS:      GET /api/v1/stats [protected]
// WHY IT EXISTS:  The dashboard overview page shows total counts:
//                 projects, skills, messages (total + unread),
//                 and experience entries.
// DEPENDS ON:     gorm, utils/response.go, models
// LAST UPDATED:   2026-05-07 — initial creation
// ============================================================

package handlers

import (
	"imtiaz-portfolio/internal/models"
	"imtiaz-portfolio/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// StatsHandler holds a direct DB reference for aggregate count queries.
type StatsHandler struct {
	db *gorm.DB
}

// NewStatsHandler creates a new StatsHandler.
func NewStatsHandler(db *gorm.DB) *StatsHandler {
	return &StatsHandler{db: db}
}

// GetStats returns aggregate counts for the dashboard overview cards.
func (h *StatsHandler) GetStats(c *gin.Context) {
	var totalProjects, totalSkills, totalExperience, totalServices int64
	var totalMessages, unreadMessages int64

	h.db.Model(&models.Project{}).Count(&totalProjects)
	h.db.Model(&models.Skill{}).Count(&totalSkills)
	h.db.Model(&models.Experience{}).Count(&totalExperience)
	h.db.Model(&models.Service{}).Count(&totalServices)
	h.db.Model(&models.Message{}).Count(&totalMessages)
	h.db.Model(&models.Message{}).Where("is_read = ?", false).Count(&unreadMessages)

	utils.Success(c, "Dashboard stats fetched", gin.H{
		"total_projects":   totalProjects,
		"total_skills":     totalSkills,
		"total_experience": totalExperience,
		"total_services":   totalServices,
		"total_messages":   totalMessages,
		"unread_messages":  unreadMessages,
	})
}
