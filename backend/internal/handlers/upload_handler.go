// ============================================================
// FILE: internal/handlers/upload_handler.go
// WHAT IT IS:     HTTP handler for file upload endpoint
// ENDPOINTS:      POST /api/v1/upload [protected]
// DEPENDS ON:     services/upload_service.go, utils/response.go
// LAST UPDATED:   2026-05-07 — initial creation
// ============================================================

package handlers

import (
	"imtiaz-portfolio/internal/services"
	"imtiaz-portfolio/internal/utils"

	"github.com/gin-gonic/gin"
)

// UploadHandler holds the upload service dependency.
type UploadHandler struct {
	uploadService *services.UploadService
}

// NewUploadHandler creates a new UploadHandler.
func NewUploadHandler(us *services.UploadService) *UploadHandler {
	return &UploadHandler{uploadService: us}
}

// Upload handles POST /api/v1/upload — saves a file and returns its URL.
func (h *UploadHandler) Upload(c *gin.Context) {
	// Read the "file" field from the multipart form data
	file, err := c.FormFile("file")
	if err != nil {
		utils.BadRequest(c, "No file provided — send the file as 'file' in form data", nil)
		return
	}

	// Read optional "folder" field to organize uploads (images, cv, logos)
	folder := c.DefaultPostForm("folder", "images")

	// Save the file and get back the public URL path
	url, err := h.uploadService.SaveFile(file, folder)
	if err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}

	utils.Created(c, "File uploaded successfully", gin.H{"url": url, "filename": file.Filename})
}
