// ============================================================
// FILE: internal/middleware/logger_middleware.go
// WHAT IT IS:     HTTP request logging middleware
// WHY IT EXISTS:  Logs every incoming request with method, path,
//                 status code, and response time. Essential for
//                 debugging production issues and monitoring traffic.
// ENDPOINTS:      Applied to ALL routes in routes.go
// DEPENDS ON:     github.com/gin-gonic/gin
// IF REMOVED:     No visibility into what requests the server receives
// LAST UPDATED:   2026-05-07 — initial creation
// ============================================================

package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

/**
 * FUNCTION: LoggerMiddleware
 * WHAT IT DOES:   Records each HTTP request after it completes.
 *                 Logs: timestamp, HTTP method, URL path, status code,
 *                 response time, and client IP address.
 *                 Colors the status code green (2xx), yellow (3xx/4xx),
 *                 or red (5xx) for easy scanning.
 * WHERE CALLED:   routes/routes.go — applied globally before all routes
 * RETURNS:        @returns {gin.HandlerFunc} - the logger middleware function
 * LAST UPDATED:   2026-05-07 — initial creation
 */
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// ─── Step 1: Record when the request arrived ──────────
		start := time.Now()
		path := c.Request.URL.Path

		// ─── Step 2: Let the handler process the request ──────
		// c.Next() runs all remaining middleware and the handler.
		// Everything after this line runs AFTER the response is sent.
		c.Next()

		// ─── Step 3: Calculate how long the request took ──────
		latency := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method

		// ─── Step 4: Choose a color for the status code ───────
		// ANSI color codes: green=32, yellow=33, red=31, reset=0
		// This makes the terminal output much easier to scan visually
		var statusColor string
		switch {
		case statusCode >= 500:
			statusColor = "\033[31m" // red — server errors need immediate attention
		case statusCode >= 400:
			statusColor = "\033[33m" // yellow — client errors (bad request, not found)
		case statusCode >= 200:
			statusColor = "\033[32m" // green — everything went well
		default:
			statusColor = "\033[0m" // no color
		}
		reset := "\033[0m"

		// ─── Step 5: Print the formatted log line ─────────────
		// Example output:
		// [2026-05-07 14:30:22] GET /api/v1/projects 200 12.3ms 127.0.0.1
		fmt.Printf("[%s] %s %s %s%d%s %v %s\n",
			time.Now().Format("2006-01-02 15:04:05"),
			method,
			path,
			statusColor,
			statusCode,
			reset,
			latency,
			clientIP,
		)
	}
}
