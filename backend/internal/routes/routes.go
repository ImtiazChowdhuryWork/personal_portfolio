// ============================================================
// FILE: internal/routes/routes.go
// WHAT IT IS:     All API route definitions in one place
// WHY IT EXISTS:  Keeps all URL-to-handler mappings in a single file
//                 so you can see the full API at a glance without
//                 searching through handler files
// DEPENDS ON:     All handlers, middleware, config, services
// IF REMOVED:     Server has no routes — every request returns 404
// LAST UPDATED:   2026-05-07 — serve frontend from Go so no Live Server needed
// ============================================================

package routes

import (
	"imtiaz-portfolio/config"
	"imtiaz-portfolio/internal/handlers"
	"imtiaz-portfolio/internal/middleware"
	"imtiaz-portfolio/internal/services"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

/**
 * FUNCTION: Setup
 * WHAT IT DOES:   Registers all HTTP routes on the Gin router.
 *                 Organizes routes into:
 *                 - Public routes (no auth needed)
 *                 - Protected routes (JWT required)
 *                 - Static file serving for uploads
 *                 - CORS and logging applied globally
 * WHERE CALLED:   cmd/main.go at server startup
 * PARAMETERS:     @param {*gin.Engine} router - the main Gin HTTP router
 *                 @param {*gorm.DB} db - database connection passed to handlers
 *                 @param {*config.Config} cfg - app config for middleware
 * LAST UPDATED:   2026-05-07 — initial creation
 */
func Setup(router *gin.Engine, db *gorm.DB, cfg *config.Config) {
	// ─── Step 1: Apply global middleware ──────────────────────
	// These run for EVERY request before any handler
	router.Use(middleware.LoggerMiddleware())
	router.Use(middleware.CORSMiddleware(cfg))

	// ─── Step 2: Serve the frontend HTML/CSS/JS files ────────
	// This means visiting http://localhost:8080 opens the portfolio directly.
	// No separate dev server (Live Server, Vite, etc.) is needed — the Go
	// server serves both the API and the frontend from the same port.
	//
	// The frontend folder is two levels up from the backend binary:
	//   backend/cmd/main.go → ../../frontend/public
	frontendPath, _ := filepath.Abs("../frontend")
	publicPath := filepath.Join(frontendPath, "public")
	assetsPath := filepath.Join(frontendPath, "assets")

	// Serve CSS, JS, and image files at /assets/*
	router.Static("/assets", assetsPath)

	// Serve uploaded user files (profile photo, CV, app screenshots)
	uploadsPath, _ := filepath.Abs(cfg.UploadDir)
	router.Static("/uploads", uploadsPath)

	// Serve the three HTML pages at clean URLs
	router.StaticFile("/", filepath.Join(publicPath, "index.html"))
	router.StaticFile("/index.html", filepath.Join(publicPath, "index.html"))
	router.StaticFile("/login", filepath.Join(publicPath, "login.html"))
	router.StaticFile("/login.html", filepath.Join(publicPath, "login.html"))
	router.StaticFile("/dashboard", filepath.Join(publicPath, "dashboard.html"))
	router.StaticFile("/dashboard.html", filepath.Join(publicPath, "dashboard.html"))

	// Health check — lets monitoring tools verify the server is alive
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy", "service": "imtiaz-portfolio-api"})
	})

	// ─── Step 3: Initialize all services with the DB connection ─
	// Services hold all business logic and database operations
	authSvc := services.NewAuthService(db, cfg)
	projectSvc := services.NewProjectService(db)
	skillSvc := services.NewSkillService(db)
	expSvc := services.NewExperienceService(db)
	msgSvc := services.NewMessageService(db)
	profileSvc := services.NewProfileService(db)
	uploadSvc := services.NewUploadService(cfg)

	// ─── Step 4: Initialize all handlers with their services ──
	// Handlers handle HTTP — they read requests and call services
	authHandler := handlers.NewAuthHandler(authSvc)
	projectHandler := handlers.NewProjectHandler(projectSvc)
	skillHandler := handlers.NewSkillHandler(skillSvc)
	expHandler := handlers.NewExperienceHandler(expSvc)
	msgHandler := handlers.NewMessageHandler(msgSvc)
	profileHandler := handlers.NewProfileHandler(profileSvc)
	uploadHandler := handlers.NewUploadHandler(uploadSvc)
	statsHandler := handlers.NewStatsHandler(db)

	// ─── Step 5: API version prefix ───────────────────────────
	// All routes are under /api/v1/ — adding v2 later won't break existing clients
	v1 := router.Group("/api/v1")

	// ─── Step 6: Public routes (no authentication needed) ─────
	// Anyone can call these to load portfolio data

	// Auth — login is public so the admin can obtain a token
	v1.POST("/auth/login", authHandler.Login)
	v1.POST("/auth/logout", authHandler.Logout)

	// Public data endpoints (the portfolio reads these on page load)
	v1.GET("/projects", projectHandler.GetAll)
	v1.GET("/projects/:id", projectHandler.GetByID)
	v1.GET("/skills", skillHandler.GetAll)
	v1.GET("/experience", expHandler.GetAll)
	v1.GET("/profile", profileHandler.Get)

	// Contact form — public so visitors can submit without logging in
	v1.POST("/messages", msgHandler.Create)

	// ─── Step 7: Protected routes (JWT required) ──────────────
	// The AuthMiddleware checks the Authorization: Bearer <token> header.
	// Any request without a valid token gets 401 Unauthorized.
	protected := v1.Group("")
	protected.Use(middleware.AuthMiddleware(cfg))

	// Auth verification — used by dashboard to confirm token is valid
	protected.GET("/auth/me", authHandler.Me)

	// Project management (dashboard CRUD)
	protected.POST("/projects", projectHandler.Create)
	protected.PUT("/projects/:id", projectHandler.Update)
	protected.DELETE("/projects/:id", projectHandler.Delete)

	// Skill management (dashboard CRUD)
	protected.POST("/skills", skillHandler.Create)
	protected.PUT("/skills/:id", skillHandler.Update)
	protected.DELETE("/skills/:id", skillHandler.Delete)

	// Experience management (dashboard CRUD)
	protected.POST("/experience", expHandler.Create)
	protected.PUT("/experience/:id", expHandler.Update)
	protected.DELETE("/experience/:id", expHandler.Delete)

	// Message management (dashboard inbox)
	protected.GET("/messages", msgHandler.GetAll)
	protected.PUT("/messages/:id", msgHandler.MarkAsRead)
	protected.DELETE("/messages/:id", msgHandler.Delete)

	// Profile editing (dashboard settings)
	protected.PUT("/profile", profileHandler.Update)

	// File uploads (images, CV PDF — dashboard only)
	protected.POST("/upload", uploadHandler.Upload)

	// Dashboard statistics overview card counts
	protected.GET("/stats", statsHandler.GetStats)
}
