package handlers

import (
	"context"
	"imtiaz-portfolio/internal/services"
	"imtiaz-portfolio/internal/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// JobHandler handles HTTP requests for job post management.
type JobHandler struct {
	jobSvc     *services.JobService
	scraperSvc *services.JobScraperService
}

// NewJobHandler creates a new JobHandler.
func NewJobHandler(jobSvc *services.JobService, scraperSvc *services.JobScraperService) *JobHandler {
	return &JobHandler{jobSvc: jobSvc, scraperSvc: scraperSvc}
}

// GetJobs handles GET /api/v1/jobs
// Query params: source, q, bookmarked, applied
func (h *JobHandler) GetJobs(c *gin.Context) {
	source := c.Query("source")
	q := c.Query("q")
	bookmarked := c.Query("bookmarked") == "true"
	applied := c.Query("applied") == "true"

	jobs, err := h.jobSvc.GetAll(source, q, bookmarked, applied)
	if err != nil {
		utils.InternalError(c, "Failed to fetch jobs")
		return
	}
	utils.Success(c, "Jobs fetched successfully", jobs)
}

// TriggerScrape handles POST /api/v1/jobs/scrape
// Launches the scraper in the background and returns immediately.
func (h *JobHandler) TriggerScrape(c *gin.Context) {
	go h.scraperSvc.ScrapeAll(context.Background())
	utils.Success(c, "Scraping started in background", nil)
}

// ToggleBookmark handles PUT /api/v1/jobs/:id/bookmark
func (h *JobHandler) ToggleBookmark(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid job ID", nil)
		return
	}
	if err := h.jobSvc.ToggleBookmark(uint(id)); err != nil {
		utils.NotFound(c, err.Error())
		return
	}
	utils.Success(c, "Bookmark updated", nil)
}

// MarkApplied handles PUT /api/v1/jobs/:id/apply
func (h *JobHandler) MarkApplied(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid job ID", nil)
		return
	}
	if err := h.jobSvc.ToggleApplied(uint(id)); err != nil {
		utils.NotFound(c, err.Error())
		return
	}
	utils.Success(c, "Applied status updated", nil)
}

// GetStats handles GET /api/v1/jobs/stats
func (h *JobHandler) GetStats(c *gin.Context) {
	stats, err := h.jobSvc.GetStats()
	if err != nil {
		utils.InternalError(c, "Failed to fetch stats")
		return
	}
	utils.Success(c, "Stats fetched", stats)
}

// UpdateStatus handles PUT /api/v1/jobs/:id/status
func (h *JobHandler) UpdateStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid job ID", nil)
		return
	}
	var body struct {
		Status      string  `json:"status"`
		AppliedAt   *string `json:"applied_at"`
		InterviewAt *string `json:"interview_at"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.BadRequest(c, "Invalid request body", err.Error())
		return
	}
	var appliedAt, interviewAt *time.Time
	if body.AppliedAt != nil && *body.AppliedAt != "" {
		if t, err := time.Parse(time.RFC3339, *body.AppliedAt); err == nil {
			appliedAt = &t
		}
	}
	if body.InterviewAt != nil && *body.InterviewAt != "" {
		if t, err := time.Parse(time.RFC3339, *body.InterviewAt); err == nil {
			interviewAt = &t
		}
	}
	if err := h.jobSvc.UpdateStatus(uint(id), body.Status, appliedAt, interviewAt); err != nil {
		utils.NotFound(c, err.Error())
		return
	}
	utils.Success(c, "Status updated", nil)
}

// UpdateNotes handles PUT /api/v1/jobs/:id/notes
func (h *JobHandler) UpdateNotes(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid job ID", nil)
		return
	}
	var body struct {
		Notes string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.BadRequest(c, "Invalid request body", err.Error())
		return
	}
	if err := h.jobSvc.UpdateNotes(uint(id), body.Notes); err != nil {
		utils.NotFound(c, err.Error())
		return
	}
	utils.Success(c, "Notes saved", nil)
}

// UpdateRecruiter handles PUT /api/v1/jobs/:id/recruiter
func (h *JobHandler) UpdateRecruiter(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid job ID", nil)
		return
	}
	var body struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		LinkedIn string `json:"linkedin"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.BadRequest(c, "Invalid request body", err.Error())
		return
	}
	if err := h.jobSvc.UpdateRecruiter(uint(id), body.Name, body.Email, body.LinkedIn); err != nil {
		utils.NotFound(c, err.Error())
		return
	}
	utils.Success(c, "Recruiter saved", nil)
}

// GetScrapeStatus handles GET /api/v1/jobs/status
// Returns the latest scrape log per source combined with configured/unconfigured state.
func (h *JobHandler) GetScrapeStatus(c *gin.Context) {
	logs, err := h.jobSvc.GetLatestScrapeLogs()
	if err != nil {
		utils.InternalError(c, "Failed to fetch scrape status")
		return
	}

	logMap := make(map[string]interface{})
	for _, l := range logs {
		logMap[l.Source] = l
	}

	sources := h.scraperSvc.SourceList()
	type row struct {
		Name       string      `json:"name"`
		Configured bool        `json:"configured"`
		LastLog    interface{} `json:"last_log"`
	}
	result := make([]row, len(sources))
	for i, s := range sources {
		name := s["name"].(string)
		result[i] = row{
			Name:       name,
			Configured: s["configured"].(bool),
			LastLog:    logMap[name],
		}
	}
	utils.Success(c, "Scrape status", result)
}

// GetScraperSettings handles GET /api/v1/jobs/settings
// Returns configured/unconfigured status only — never exposes raw key values.
func (h *JobHandler) GetScraperSettings(c *gin.Context) {
	row, err := h.jobSvc.GetScraperSettings()
	if err != nil {
		utils.InternalError(c, "Failed to load settings")
		return
	}
	// Defaults when no row exists yet
	adzunaID, adzunaKey, rapidKey := "", "", ""
	interval, maxAge, alertMin := 6, 7, 5
	disabled := ""
	alertEnabled := false
	if row != nil {
		adzunaID     = row.AdzunaAppID
		adzunaKey    = row.AdzunaAppKey
		rapidKey     = row.RapidAPIKey
		disabled     = row.DisabledSources
		alertEnabled = row.AlertEnabled
		if row.ScrapeIntervalHours > 0 {
			interval = row.ScrapeIntervalHours
		}
		if row.MaxAgeDays > 0 {
			maxAge = row.MaxAgeDays
		}
		if row.AlertMinNewJobs > 0 {
			alertMin = row.AlertMinNewJobs
		}
	}
	utils.Success(c, "Settings loaded", gin.H{
		"adzuna_configured":    adzunaID != "" && adzunaKey != "",
		"jsearch_configured":   rapidKey != "",
		"adzuna_app_id":        adzunaID,
		"adzuna_app_key":       adzunaKey,
		"rapidapi_key":         rapidKey,
		"scrape_interval_hours": interval,
		"max_age_days":          maxAge,
		"disabled_sources":      disabled,
		"alert_enabled":         alertEnabled,
		"alert_min_new_jobs":    alertMin,
	})
}

// SaveScraperSettings handles PUT /api/v1/jobs/settings
func (h *JobHandler) SaveScraperSettings(c *gin.Context) {
	var req services.SaveScraperSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Invalid request body", err.Error())
		return
	}
	if err := h.jobSvc.SaveScraperSettings(req); err != nil {
		utils.InternalError(c, "Failed to save settings")
		return
	}
	utils.Success(c, "Settings saved", nil)
}

// GetKeywords handles GET /api/v1/jobs/keywords
func (h *JobHandler) GetKeywords(c *gin.Context) {
	keywords, err := h.jobSvc.GetKeywords()
	if err != nil {
		utils.InternalError(c, "Failed to fetch keywords")
		return
	}
	utils.Success(c, "Keywords fetched", keywords)
}

// SetKeywords handles PUT /api/v1/jobs/keywords
func (h *JobHandler) SetKeywords(c *gin.Context) {
	var body struct {
		Keywords []string `json:"keywords"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.BadRequest(c, "Invalid request body", err.Error())
		return
	}
	if err := h.jobSvc.SetKeywords(body.Keywords); err != nil {
		utils.InternalError(c, "Failed to save keywords")
		return
	}
	utils.Success(c, "Keywords saved", nil)
}

// DeleteJob handles DELETE /api/v1/jobs/:id
func (h *JobHandler) DeleteJob(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.BadRequest(c, "Invalid job ID", nil)
		return
	}
	if err := h.jobSvc.DeleteJob(uint(id)); err != nil {
		utils.NotFound(c, err.Error())
		return
	}
	utils.Success(c, "Job deleted", nil)
}
