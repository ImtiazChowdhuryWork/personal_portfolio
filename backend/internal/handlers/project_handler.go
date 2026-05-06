// ============================================================
// FILE: internal/handlers/project_handler.go
// WHAT IT IS:     HTTP handlers for the projects/apps API
// WHY IT EXISTS:  Handles GET (public) and POST/PUT/DELETE (protected)
//                 requests for the Apps Showcase data
// ENDPOINTS:      GET    /api/v1/projects
//                 GET    /api/v1/projects/:id
//                 POST   /api/v1/projects    [protected]
//                 PUT    /api/v1/projects/:id [protected]
//                 DELETE /api/v1/projects/:id [protected]
// DEPENDS ON:     services/project_service.go, utils/response.go
// IF REMOVED:     Apps Showcase section cannot load or be managed
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

// ProjectHandler holds the project service dependency.
type ProjectHandler struct {
	projectService *services.ProjectService
}

// NewProjectHandler creates a new ProjectHandler.
func NewProjectHandler(ps *services.ProjectService) *ProjectHandler {
	return &ProjectHandler{projectService: ps}
}

// GetAll handles GET /api/v1/projects — returns all projects (public)
func (h *ProjectHandler) GetAll(c *gin.Context) {
	projects, err := h.projectService.GetAll()
	if err != nil {
		utils.InternalError(c, "Failed to fetch projects")
		return
	}
	utils.Success(c, "Projects fetched successfully", projects)
}

// GetByID handles GET /api/v1/projects/:id — returns a single project (public)
func (h *ProjectHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid project ID", nil)
		return
	}
	project, err := h.projectService.GetByID(uint(id))
	if err != nil {
		utils.NotFound(c, err.Error())
		return
	}
	utils.Success(c, "Project fetched", project)
}

// Create handles POST /api/v1/projects — creates a new project [protected]
func (h *ProjectHandler) Create(c *gin.Context) {
	var project models.Project
	if err := c.ShouldBindJSON(&project); err != nil {
		utils.BadRequest(c, "Invalid project data", err.Error())
		return
	}
	if utils.IsEmpty(project.Name) {
		utils.BadRequest(c, "Project name is required", nil)
		return
	}
	if err := h.projectService.Create(&project); err != nil {
		utils.InternalError(c, "Failed to create project")
		return
	}
	utils.Created(c, "Project created successfully", project)
}

// Update handles PUT /api/v1/projects/:id — updates a project [protected]
func (h *ProjectHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid project ID", nil)
		return
	}
	var updates models.Project
	if err := c.ShouldBindJSON(&updates); err != nil {
		utils.BadRequest(c, "Invalid project data", err.Error())
		return
	}
	project, err := h.projectService.Update(uint(id), &updates)
	if err != nil {
		utils.NotFound(c, err.Error())
		return
	}
	utils.Success(c, "Project updated successfully", project)
}

// Delete handles DELETE /api/v1/projects/:id — deletes a project [protected]
func (h *ProjectHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid project ID", nil)
		return
	}
	if err := h.projectService.Delete(uint(id)); err != nil {
		utils.NotFound(c, err.Error())
		return
	}
	utils.Success(c, "Project deleted successfully", nil)
}
