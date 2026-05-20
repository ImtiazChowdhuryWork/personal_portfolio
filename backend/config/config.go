// ============================================================
// FILE: config/config.go
// WHAT IT IS:     Application configuration loader
// WHY IT EXISTS:  Centralizes all environment variable reading
//                 so the rest of the app never calls os.Getenv
//                 directly — one place to change config behavior
// ENDPOINTS:      N/A — internal package only
// DEPENDS ON:     github.com/joho/godotenv, os package
// IF REMOVED:     Entire app loses its configuration — nothing works
// LAST UPDATED:   2026-05-07 — initial creation
// ============================================================

package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	// godotenv reads the .env file and loads variables into the process environment
	// Without this, you'd have to set env vars manually before running the server
	"github.com/joho/godotenv"
)

// Config holds all application settings loaded from environment variables.
// Every other package in the app receives a pointer to this struct
// so they never need to read env vars themselves.
type Config struct {
	// ─── Server ──────────────────────────────────────────────
	ServerPort string // The port the HTTP server binds to (e.g. "8080")
	AppEnv     string // "development" or "production" — affects error detail

	// ─── Database ────────────────────────────────────────────
	DBHost     string // PostgreSQL server hostname
	DBPort     string // PostgreSQL server port
	DBUser     string // Database username
	DBPassword string // Database password
	DBName     string // Which database to connect to
	DBSSLMode  string // "disable" for local, "require" for production

	// ─── JWT ─────────────────────────────────────────────────
	JWTSecret         string // Secret key for signing tokens — must stay private
	JWTExpiry         string // How long access tokens live (e.g. "24h")
	JWTRefreshExpiry  string // How long refresh tokens live (e.g. "168h")

	// ─── File Uploads ────────────────────────────────────────
	UploadDir       string // Directory path where uploaded files are saved
	MaxUploadSizeMB int64  // Maximum allowed upload size in megabytes

	// ─── CORS ────────────────────────────────────────────────
	AllowedOrigins string // Comma-separated list of allowed frontend origins

	// ─── Admin ───────────────────────────────────────────────
	AdminEmail    string // Default admin email created on first boot
	AdminPassword string // Default admin password (should be changed after login)

	// ─── SMTP (email sending) ────────────────────────────────
	SMTPHost string // e.g. smtp.gmail.com
	SMTPPort string // e.g. 587
	SMTPUser string // Gmail address
	SMTPPass string // Gmail App Password (not regular password)

	// ─── Job Scraper ─────────────────────────────────────────
	JobSearchKeywords string // comma-separated, e.g. "flutter,dart,mobile developer"
	AdzunaAppID       string // free at developer.adzuna.com
	AdzunaAppKey      string
	RapidAPIKey       string // free at rapidapi.com — enables JSearch (LinkedIn+Indeed+Glassdoor)
}

/**
 * FUNCTION: Load
 * WHAT IT DOES:   Reads the .env file, then reads each environment
 *                 variable into a Config struct and returns it.
 *                 If any required variable is missing, it uses a
 *                 sensible default so the app can still start.
 * WHY IT EXISTS:  Centralizes all config access — nothing else
 *                 should ever call os.Getenv
 * WHERE CALLED:   cmd/main.go → at application startup
 * RETURNS:        @returns {*Config} - pointer to filled config struct
 * IF CHANGED:     All handlers and services receive config via
 *                 dependency injection, so adding a field here
 *                 requires also passing it where needed
 * LAST UPDATED:   2026-05-07 — initial creation
 */
func Load() *Config {
	// ─── Step 1: Load the .env file ──────────────────────────
	// godotenv.Load reads .env and sets each KEY=VALUE as an
	// actual OS environment variable. We ignore the error here
	// because in production the vars may already be set by the
	// host environment (Railway, Render, etc.) without a .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found — using system environment variables")
	}

	// ─── Step 2: Parse MAX_UPLOAD_SIZE_MB as integer ─────────
	// os.Getenv always returns a string, so we convert to int64
	// Default to 10 MB if the variable is missing or not a number
	maxUploadMB, err := strconv.ParseInt(getEnv("MAX_UPLOAD_SIZE_MB", "10"), 10, 64)
	if err != nil {
		// If the value in .env is not a valid number, fall back to 10
		maxUploadMB = 10
	}

	// ─── Step 3: Build and return the config struct ───────────
	// getEnv(key, default) returns the env var value or the default
	// if the variable is empty or not set
	return &Config{
		ServerPort:       getEnv("SERVER_PORT", "8080"),
		AppEnv:           getEnv("APP_ENV", "development"),
		DBHost:           getEnv("DB_HOST", "localhost"),
		DBPort:           getEnv("DB_PORT", "5432"),
		DBUser:           getEnv("DB_USER", "postgres"),
		DBPassword:       getEnv("DB_PASSWORD", "postgres"),
		DBName:           getEnv("DB_NAME", "imtiaz_portfolio"),
		DBSSLMode:        getEnv("DB_SSLMODE", "disable"),
		JWTSecret:        getEnv("JWT_SECRET", "fallback-secret-change-me"),
		JWTExpiry:        getEnv("JWT_EXPIRY", "24h"),
		JWTRefreshExpiry: getEnv("JWT_REFRESH_EXPIRY", "168h"),
		UploadDir:        getEnv("UPLOAD_DIR", "./uploads"),
		MaxUploadSizeMB:  maxUploadMB,
		AllowedOrigins:   getEnv("ALLOWED_ORIGINS", "http://localhost:3000"),
		AdminEmail:       getEnv("ADMIN_EMAIL", "admin@imtiaz.dev"),
		AdminPassword:    getEnv("ADMIN_PASSWORD", "Admin@1234"),
		SMTPHost:         getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:         getEnv("SMTP_PORT", "587"),
		SMTPUser:         getEnv("SMTP_USER", ""),
		SMTPPass:         getEnv("SMTP_PASS", ""),
		JobSearchKeywords: getEnv("JOB_SEARCH_KEYWORDS", "flutter,dart,mobile developer"),
		AdzunaAppID:       getEnv("ADZUNA_APP_ID", ""),
		AdzunaAppKey:      getEnv("ADZUNA_APP_KEY", ""),
		RapidAPIKey:       getEnv("RAPIDAPI_KEY", ""),
	}
}

/**
 * FUNCTION: DSN
 * WHAT IT DOES:   Builds the PostgreSQL Data Source Name string
 *                 from the individual DB config fields.
 *                 GORM needs this exact format to connect.
 * WHY IT EXISTS:  Keeps DSN construction in one place so if the
 *                 format ever changes, only this function updates
 * WHERE CALLED:   config/database.go → ConnectDB
 * RETURNS:        @returns {string} - full PostgreSQL DSN string
 * EXAMPLE:        cfg.DSN() → "host=localhost user=postgres ..."
 * LAST UPDATED:   2026-05-07 — initial creation
 */
func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Dhaka",
		c.DBHost,
		c.DBUser,
		c.DBPassword,
		c.DBName,
		c.DBPort,
		c.DBSSLMode,
	)
}

/**
 * FUNCTION: getEnv
 * WHAT IT DOES:   Returns the value of an environment variable,
 *                 or a default value if the variable is not set
 *                 or is an empty string
 * WHY IT EXISTS:  Prevents repeated os.Getenv + empty check patterns
 *                 all over the codebase
 * WHERE CALLED:   Load() — for every config field
 * PARAMETERS:     @param {string} key - the environment variable name
 *                 @param {string} defaultVal - fallback if not set
 * RETURNS:        @returns {string} - env value or default
 * LAST UPDATED:   2026-05-07 — initial creation
 */
func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
