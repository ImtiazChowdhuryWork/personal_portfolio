// ============================================================
// FILE: internal/services/project_service.go
// WHAT IT IS:     Project/App CRUD business logic
// WHY IT EXISTS:  All database operations for the Apps Showcase
//                 section live here. Keeps handlers thin and testable.
// ENDPOINTS:      Called by project_handler.go
// DEPENDS ON:     models/project.go, gorm
// IF REMOVED:     Project handler has no database operations
// LAST UPDATED:   2026-05-07 — initial creation
// ============================================================

package services

import (
	"errors"
	"imtiaz-portfolio/internal/models"
	"strings"

	"gorm.io/gorm"
)

// ProjectService handles all database operations for projects.
type ProjectService struct {
	db *gorm.DB
}

// NewProjectService creates a new ProjectService with a database connection.
func NewProjectService(db *gorm.DB) *ProjectService {
	return &ProjectService{db: db}
}

/**
 * FUNCTION: GetAll
 * WHAT IT DOES:   Fetches all projects ordered by SortOrder then CreatedAt.
 *                 Featured projects appear first in the showcase slider.
 * WHERE CALLED:   project_handler.go → GET /api/v1/projects
 * LAST UPDATED:   2026-05-07 — initial creation
 */
func (s *ProjectService) GetAll() ([]models.Project, error) {
	var projects []models.Project
	result := s.db.Order("sort_order ASC, created_at DESC").Find(&projects)
	return projects, result.Error
}

/**
 * FUNCTION: GetByID
 * WHAT IT DOES:   Fetches a single project by its database ID.
 * WHERE CALLED:   project_handler.go → GET /api/v1/projects/:id
 * LAST UPDATED:   2026-05-07 — initial creation
 */
func (s *ProjectService) GetByID(id uint) (*models.Project, error) {
	var project models.Project
	result := s.db.First(&project, id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, errors.New("project not found")
	}
	return &project, result.Error
}

/**
 * FUNCTION: Create
 * WHAT IT DOES:   Inserts a new project row into the database.
 *                 Auto-generates a URL slug from the project name.
 * WHERE CALLED:   project_handler.go → POST /api/v1/projects
 * LAST UPDATED:   2026-05-07 — initial creation
 */
func (s *ProjectService) Create(project *models.Project) error {
	// Auto-generate slug from name if not provided
	if project.Slug == "" {
		project.Slug = generateSlug(project.Name)
	}
	return s.db.Create(project).Error
}

/**
 * FUNCTION: Update
 * WHAT IT DOES:   Updates an existing project's fields in the database.
 *                 Only updates fields that are provided (partial update).
 * WHERE CALLED:   project_handler.go → PUT /api/v1/projects/:id
 * LAST UPDATED:   2026-05-07 — initial creation
 */
func (s *ProjectService) Update(id uint, updates *models.Project) (*models.Project, error) {
	var project models.Project
	if err := s.db.First(&project, id).Error; err != nil {
		return nil, errors.New("project not found")
	}
	s.db.Model(&project).Updates(updates)
	return &project, nil
}

/**
 * FUNCTION: Delete
 * WHAT IT DOES:   Soft-deletes a project (sets DeletedAt timestamp).
 *                 The row stays in the database but won't appear in queries.
 *                 Soft delete prevents accidental permanent data loss.
 * WHERE CALLED:   project_handler.go → DELETE /api/v1/projects/:id
 * LAST UPDATED:   2026-05-07 — initial creation
 */
func (s *ProjectService) Delete(id uint) error {
	result := s.db.Delete(&models.Project{}, id)
	if result.RowsAffected == 0 {
		return errors.New("project not found")
	}
	return result.Error
}

// generateSlug converts a name like "Project Finder" into "project-finder"
// for use in URLs. Replaces spaces with hyphens and lowercases everything.
func generateSlug(name string) string {
	slug := strings.ToLower(strings.TrimSpace(name))
	slug = strings.ReplaceAll(slug, " ", "-")
	// Remove characters that are not alphanumeric or hyphens
	var result strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	return result.String()
}
