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
	serviceSvc := services.NewServiceService(db)
	msgSvc := services.NewMessageService(db)
	emailSvc := services.NewEmailService(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass)
	profileSvc := services.NewProfileService(db)
	uploadSvc := services.NewUploadService(cfg)
	cvGenSvc := services.NewCVGeneratorService(db, profileSvc, cfg.UploadDir)
	// Real-time notification broker — every connected dashboard subscribes
	// via /api/v1/events and gets pushed when new messages arrive.
	broker := services.NewEventBroker()

	// Async reply queue — Reply handler enqueues here, a worker goroutine
	// inside the queue does the actual SMTP send and reports back through
	// the broker so the dashboard can flip ticks in real time.
	replyQueue := services.NewReplyQueue(emailSvc, msgSvc, profileSvc, broker, cfg.UploadDir)

	// Recover any replies that were left "pending" by a previous run (e.g.
	// the server was killed mid-send). Worker will retry them now.
	if pending, err := msgSvc.GetPendingReplies(); err == nil {
		for _, r := range pending {
			origMsg, err := msgSvc.GetByID(r.MessageID)
			if err != nil {
				continue
			}
			subject := "Re: " + origMsg.Type + " — Imtiaz Portfolio"
			fullBody := r.ReplyText + "\n\n---\nOriginal message from " + origMsg.Name + ":\n" + origMsg.Content
			replyQueue.Enqueue(services.ReplyJob{
				ReplyID:        r.ID,
				ToName:         origMsg.Name,
				ToEmail:        origMsg.Email,
				FromEmail:      r.FromEmail,
				Subject:        subject,
				Body:           fullBody,
				AttachmentURLs: r.Attachments,
			})
		}
	}

	// ─── Step 4: Initialize all handlers with their services ──
	// Handlers handle HTTP — they read requests and call services
	authHandler := handlers.NewAuthHandler(authSvc)
	projectHandler := handlers.NewProjectHandler(projectSvc)
	skillHandler := handlers.NewSkillHandler(skillSvc)
	expHandler := handlers.NewExperienceHandler(expSvc)
	serviceHandler := handlers.NewServiceHandler(serviceSvc)
	msgHandler := handlers.NewMessageHandler(msgSvc, emailSvc, profileSvc, broker, replyQueue, cfg.UploadDir)
	profileHandler := handlers.NewProfileHandler(profileSvc, emailSvc, cvGenSvc, cfg.UploadDir)
	uploadHandler := handlers.NewUploadHandler(uploadSvc)
	statsHandler := handlers.NewStatsHandler(db)
	sseHandler := handlers.NewSSEHandler(broker, cfg)

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
	v1.GET("/services", serviceHandler.GetAll)
	v1.GET("/profile", profileHandler.Get)

	// Contact form — public so visitors can submit without logging in
	v1.POST("/messages", msgHandler.Create)

	// Public CV download counter — fired by the portfolio when a visitor
	// clicks "Download CV", just before navigating to the PDF
	v1.POST("/profile/cv/download", profileHandler.RecordCVDownload)

	// Real-time event stream for the dashboard. Auth lives inside the
	// handler (?token=…) because EventSource cannot send headers; the
	// route itself stays in the public group so the middleware doesn't
	// reject it before the handler can read the query.
	v1.GET("/events", sseHandler.Stream)

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

	// Services / What I Offer (dashboard CRUD)
	protected.GET("/services/admin", serviceHandler.GetAllAdmin)
	protected.POST("/services", serviceHandler.Create)
	protected.PUT("/services/:id", serviceHandler.Update)
	protected.DELETE("/services/:id", serviceHandler.Delete)

	// Message management (dashboard inbox)
	protected.GET("/messages", msgHandler.GetAll)
	protected.PUT("/messages/:id", msgHandler.MarkAsRead)
	protected.POST("/messages/:id/reply", msgHandler.Reply)
	protected.GET("/messages/:id/replies", msgHandler.GetReplies)
	protected.DELETE("/messages/:id", msgHandler.Delete)

	// Re-send a previously-failed reply via the same async queue.
	protected.POST("/messages/replies/:id/retry", msgHandler.RetryReply)

	// Profile editing (dashboard settings)
	protected.PUT("/profile", profileHandler.Update)
	protected.PUT("/profile/cv", profileHandler.UpdateCV)
	protected.PUT("/profile/cv/visibility", profileHandler.UpdateCVVisibility)
	protected.POST("/profile/cv/generate", profileHandler.GenerateCV)
	protected.GET("/profile/cv/history", profileHandler.GetCVHistory)
	protected.POST("/profile/cv/history/:id/activate", profileHandler.ActivateCV)
	protected.DELETE("/profile/cv/history/:id", profileHandler.DeleteCVHistory)
	protected.PUT("/profile/photo", profileHandler.UpdatePhoto)
	protected.PUT("/profile/about-photo", profileHandler.UpdateAboutPhoto)
	protected.PUT("/profile/social", profileHandler.UpdateSocial)
	protected.PUT("/profile/mail", profileHandler.UpdateMail)
	protected.PUT("/profile/footer", profileHandler.UpdateFooter)
	protected.PUT("/profile/github", profileHandler.UpdateGitHub)
	protected.GET("/profile/mail/history", profileHandler.GetMailHistory)
	protected.DELETE("/profile/mail/history/:id", profileHandler.DeleteMailHistory)
	protected.POST("/profile/mail/accounts/:id/hide", profileHandler.HideMailAccount)
	protected.POST("/profile/mail/accounts/:id/unhide", profileHandler.UnhideMailAccount)
	protected.POST("/profile/mail/verify", profileHandler.VerifyMail)

	// File uploads (images, CV PDF — dashboard only)
	protected.POST("/upload", uploadHandler.Upload)

	// Dashboard statistics overview card counts
	protected.GET("/stats", statsHandler.GetStats)
}
