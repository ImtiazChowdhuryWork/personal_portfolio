// ============================================================
// FILE: cmd/main.go
// WHAT IT IS:     Application entry point — the main function
// WHY IT EXISTS:  This is where everything starts. It loads config,
//                 connects to the database, runs migrations, seeds
//                 default data, sets up routes, and starts the HTTP server.
// DEPENDS ON:     config, gorm, routes, seeds, gin
// IF REMOVED:     The entire application cannot start
// LAST UPDATED:   2026-05-07 — initial creation
// ============================================================

package main

import (
	"fmt"
	"imtiaz-portfolio/config"
	"imtiaz-portfolio/internal/models"
	"imtiaz-portfolio/internal/routes"
	"imtiaz-portfolio/seeds"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Println("🚀 Starting Imtiaz Portfolio API Server...")
	fmt.Println("==========================================")

	// ─── Step 1: Load configuration from .env ─────────────────
	// All settings come from environment variables — never hardcoded
	cfg := config.Load()
	fmt.Printf("   Environment: %s\n", cfg.AppEnv)
	fmt.Printf("   Server Port: %s\n", cfg.ServerPort)

	// ─── Step 2: Set Gin mode based on environment ─────────────
	// In production, Gin's release mode disables debug logs that
	// could expose internal routes and middleware to attackers
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// ─── Step 3: Connect to PostgreSQL ────────────────────────
	db := config.ConnectDB(cfg)

	// ─── Step 4: Auto-migrate database schema ─────────────────
	// GORM reads the model structs and creates/updates tables automatically.
	// This is safe to run on every startup — it only adds missing columns,
	// never drops existing ones (use proper migrations for destructive changes).
	fmt.Println("   Running database migrations...")
	if err := db.AutoMigrate(
		&models.User{},
		&models.Project{},
		&models.Skill{},
		&models.Experience{},
		&models.Testimonial{},
		&models.Message{},
		&models.MessageReply{},
		&models.Profile{},
		&models.MailPasswordHistory{},
		&models.HiddenMailAccount{},
		&models.CVFile{},
		&models.Service{},
	); err != nil {
		log.Fatal("Database migration failed: ", err)
	}
	fmt.Println("   ✓ Migrations complete")

	// ─── Step 5: Seed default data ────────────────────────────
	// Creates admin user and sample content if database is empty.
	// Safe to call on every startup — checks for existing data first.
	seeds.Run(db, cfg)

	// ─── Step 6: Create the uploads directory ─────────────────
	// Ensure the uploads directory exists before serving files from it
	if err := os.MkdirAll(cfg.UploadDir, 0755); err != nil {
		log.Printf("⚠️  Could not create uploads directory: %v", err)
	}

	// ─── Step 7: Set up the HTTP router and routes ────────────
	router := gin.New() // gin.New() skips default middleware — we add our own
	routes.Setup(router, db, cfg)

	// ─── Step 8: Configure the HTTP server ────────────────────
	// Setting timeouts prevents slow clients from holding connections open forever.
	// 5 minutes is generous enough for slow-upload reply sends with multiple
	// large attachments — at residential upload speeds, a ~15 MB email body to
	// Gmail can easily take 1–2 minutes. The SMTP service has its own 120s
	// deadline below this, so the server timeout is mainly a backstop against
	// truly stuck clients.
	server := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      router,
		ReadTimeout:  5 * time.Minute,   // max time to read the full request (covers slow uploads)
		WriteTimeout: 5 * time.Minute,   // covers handler execution + response write
		IdleTimeout:  120 * time.Second, // max time to keep idle connections open
	}

	// ─── Step 9: Start the server in a background goroutine ───
	// Running in a goroutine lets us handle shutdown signals below
	// without blocking the main thread
	go func() {
		fmt.Printf("\n✅ Server running at http://localhost:%s\n", cfg.ServerPort)
		fmt.Printf("   API base URL: http://localhost:%s/api/v1\n", cfg.ServerPort)
		fmt.Println("   Press Ctrl+C to stop")
		fmt.Println("==========================================\n")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed to start: ", err)
		}
	}()

	// ─── Step 10: Graceful shutdown ───────────────────────────
	// Listen for OS signals (Ctrl+C or SIGTERM from Docker/systemd).
	// When received, give in-flight requests up to 10 seconds to finish
	// before forcing the server to stop. This prevents cutting off
	// requests that are mid-flight (e.g. a file upload).
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // Blocks here until a signal is received

	fmt.Println("\n⏳ Shutting down gracefully...")

	// Close the database connection pool cleanly
	sqlDB, _ := db.DB()
	sqlDB.Close()

	fmt.Println("✅ Server stopped. Goodbye!")
}
