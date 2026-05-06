// ============================================================
// FILE: internal/services/upload_service.go
// WHAT IT IS:     File upload handling business logic
// WHY IT EXISTS:  Manages saving uploaded images (profile photo,
//                 app screenshots, company logos) and the CV PDF
//                 to the /uploads directory on the server.
// ENDPOINTS:      Called by upload_handler.go
// DEPENDS ON:     config, mime/multipart, os, path/filepath
// IF REMOVED:     File uploads fail — images cannot be added to portfolio
// LAST UPDATED:   2026-05-07 — initial creation
// ============================================================

package services

import (
	"errors"
	"fmt"
	"imtiaz-portfolio/config"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// UploadService handles saving files from multipart form uploads.
type UploadService struct {
	cfg *config.Config
}

// NewUploadService creates a new UploadService.
func NewUploadService(cfg *config.Config) *UploadService {
	return &UploadService{cfg: cfg}
}

// allowedImageTypes maps file extensions to whether they are allowed.
// Only image files and PDFs can be uploaded to prevent abuse.
var allowedTypes = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".webp": true,
	".svg":  true,
	".pdf":  true,
}

/**
 * FUNCTION: SaveFile
 * WHAT IT DOES:   Validates and saves an uploaded file to the uploads
 *                 directory. Generates a unique filename using a timestamp
 *                 to prevent name collisions (two users uploading "photo.jpg").
 * WHERE CALLED:   upload_handler.go → POST /api/v1/upload
 * PARAMETERS:     @param {*multipart.FileHeader} file - the uploaded file info
 *                 @param {string} subfolder - "images", "cv", "logos" etc.
 * RETURNS:        @returns {string} - the public URL path to the saved file
 *                 @returns {error} - non-nil if validation or save failed
 * LAST UPDATED:   2026-05-07 — initial creation
 */
func (s *UploadService) SaveFile(file *multipart.FileHeader, subfolder string) (string, error) {
	// ─── Step 1: Validate file size ───────────────────────────
	maxBytes := s.cfg.MaxUploadSizeMB * 1024 * 1024
	if file.Size > maxBytes {
		return "", fmt.Errorf("file size exceeds the %dMB limit", s.cfg.MaxUploadSizeMB)
	}

	// ─── Step 2: Validate file extension ──────────────────────
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedTypes[ext] {
		return "", errors.New("file type not allowed — only images and PDFs accepted")
	}

	// ─── Step 3: Create the upload directory if it doesn't exist ─
	uploadPath := filepath.Join(s.cfg.UploadDir, subfolder)
	if err := os.MkdirAll(uploadPath, 0755); err != nil {
		return "", errors.New("failed to create upload directory")
	}

	// ─── Step 4: Generate a unique filename ───────────────────
	// Using Unix timestamp + original name prevents filename collisions
	// Example: "1715078400_profile.jpg"
	uniqueName := fmt.Sprintf("%d_%s", time.Now().Unix(), filepath.Base(file.Filename))
	savePath := filepath.Join(uploadPath, uniqueName)

	// ─── Step 5: Open the uploaded file ───────────────────────
	src, err := file.Open()
	if err != nil {
		return "", errors.New("failed to open uploaded file")
	}
	defer src.Close()

	// ─── Step 6: Create the destination file ──────────────────
	dst, err := os.Create(savePath)
	if err != nil {
		return "", errors.New("failed to create destination file on server")
	}
	defer dst.Close()

	// ─── Step 7: Copy the file contents ───────────────────────
	buf := make([]byte, 32*1024) // 32KB buffer for efficient copying
	for {
		n, err := src.Read(buf)
		if n > 0 {
			if _, werr := dst.Write(buf[:n]); werr != nil {
				return "", errors.New("failed to write file to disk")
			}
		}
		if err != nil {
			break
		}
	}

	// Return the public URL path (relative to server root)
	// The frontend uses this path to display the uploaded file
	return fmt.Sprintf("/uploads/%s/%s", subfolder, uniqueName), nil
}
