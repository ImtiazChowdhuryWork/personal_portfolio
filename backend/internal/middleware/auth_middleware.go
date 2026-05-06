// ============================================================
// FILE: internal/middleware/auth_middleware.go
// WHAT IT IS:     JWT authentication middleware for protected routes
// WHY IT EXISTS:  Protected API routes (POST/PUT/DELETE projects,
//                 messages, etc.) must only work when the caller
//                 is authenticated. This middleware runs before
//                 any protected handler and rejects unauthorized calls.
// ENDPOINTS:      Applied to all [protected] routes in routes.go
// DEPENDS ON:     utils/jwt.go, utils/response.go, config/config.go
// IF REMOVED:     Anyone on the internet can delete/edit all portfolio data
// LAST UPDATED:   2026-05-07 — initial creation
// ============================================================

package middleware

import (
	"imtiaz-portfolio/config"
	"imtiaz-portfolio/internal/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

/**
 * FUNCTION: AuthMiddleware
 * WHAT IT DOES:   Returns a Gin middleware function that:
 *                 1. Reads the Authorization header
 *                 2. Extracts the Bearer token
 *                 3. Validates the JWT signature and expiry
 *                 4. If valid: stores user info in context and continues
 *                 5. If invalid: responds 401 and stops the request
 * WHY IT EXISTS:  Without this, any request could reach protected handlers
 * WHERE CALLED:   routes/routes.go — applied to the "protected" route group
 * PARAMETERS:     @param {*config.Config} cfg - holds JWT_SECRET
 * RETURNS:        @returns {gin.HandlerFunc} - the middleware function
 * EXAMPLE:        protected := router.Group("/api/v1").Use(AuthMiddleware(cfg))
 * LAST UPDATED:   2026-05-07 — initial creation
 */
func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// ─── Step 1: Read the Authorization header ────────────
		// The frontend sends: Authorization: Bearer <token>
		// If this header is missing the request is definitely unauthorized
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.Unauthorized(c, "Authorization header is required")
			// c.Abort() stops the request from reaching the actual handler
			c.Abort()
			return
		}

		// ─── Step 2: Extract the token from "Bearer <token>" ──
		// The header format is exactly: "Bearer " followed by the token
		// We split on space and check there are exactly 2 parts
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.Unauthorized(c, "Authorization header format must be: Bearer <token>")
			c.Abort()
			return
		}
		tokenString := parts[1]

		// ─── Step 3: Validate the JWT token ───────────────────
		// ValidateToken checks the signature and expiry.
		// If the token was tampered with or has expired, it returns an error.
		claims, err := utils.ValidateToken(tokenString, cfg.JWTSecret)
		if err != nil {
			utils.Unauthorized(c, "Invalid or expired token — please log in again")
			c.Abort()
			return
		}

		// ─── Step 4: Store user info in the request context ───
		// Downstream handlers can read these values with:
		//   userID := c.GetUint("userID")
		//   email := c.GetString("email")
		// This avoids having to re-parse the token in every handler
		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)

		// ─── Step 5: Continue to the actual handler ───────────
		// c.Next() passes control to the next middleware or the handler
		c.Next()
	}
}
