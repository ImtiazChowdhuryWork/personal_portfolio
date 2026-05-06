// ============================================================
// FILE: internal/middleware/cors_middleware.go
// WHAT IT IS:     CORS (Cross-Origin Resource Sharing) middleware
// WHY IT EXISTS:  Browsers block JavaScript from calling APIs on
//                 different domains by default. CORS headers tell
//                 the browser which frontends are allowed to call
//                 this API. Without this, the portfolio frontend
//                 cannot make ANY API requests.
// ENDPOINTS:      Applied to ALL routes in routes.go
// DEPENDS ON:     github.com/gin-contrib/cors, config/config.go
// IF REMOVED:     Frontend gets "CORS error" on every API call — nothing works
// LAST UPDATED:   2026-05-07 — initial creation
// ============================================================

package middleware

import (
	"imtiaz-portfolio/config"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

/**
 * FUNCTION: CORSMiddleware
 * WHAT IT DOES:   Creates and returns a CORS middleware configured
 *                 with the allowed origins from the .env file.
 *                 Allows the standard HTTP methods and headers
 *                 that the frontend JavaScript uses.
 * WHERE CALLED:   routes/routes.go — applied globally before all routes
 * PARAMETERS:     @param {*config.Config} cfg - holds ALLOWED_ORIGINS
 * RETURNS:        @returns {gin.HandlerFunc} - the configured CORS middleware
 * LAST UPDATED:   2026-05-07 — initial creation
 */
func CORSMiddleware(cfg *config.Config) gin.HandlerFunc {
	// Parse the comma-separated origins string into a slice
	// Example: "http://localhost:3000,http://127.0.0.1:5500"
	// becomes: ["http://localhost:3000", "http://127.0.0.1:5500"]
	origins := strings.Split(cfg.AllowedOrigins, ",")
	for i, o := range origins {
		origins[i] = strings.TrimSpace(o)
	}

	return cors.New(cors.Config{
		// AllowOrigins lists which frontend URLs can call the API.
		// Requests from any other domain will be blocked by the browser.
		AllowOrigins: origins,

		// AllowMethods lists which HTTP methods are permitted.
		// The frontend uses GET (read), POST (create), PUT (update),
		// DELETE (remove), and OPTIONS (preflight check by the browser).
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},

		// AllowHeaders lists which request headers are permitted.
		// Authorization is needed for the JWT token.
		// Content-Type is needed to send JSON bodies.
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Authorization",
			"Accept",
			"X-Requested-With",
		},

		// ExposeHeaders lists which response headers the frontend can read.
		ExposeHeaders: []string{"Content-Length"},

		// AllowCredentials allows cookies and Authorization headers to be sent.
		// Required for JWT auth to work across origins.
		AllowCredentials: true,

		// MaxAge tells the browser how long to cache the preflight response.
		// 12 hours means the browser won't send a preflight OPTIONS request
		// on every single API call — only once every 12 hours.
		MaxAge: 12 * time.Hour,
	})
}
