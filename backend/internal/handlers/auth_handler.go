// ============================================================
// FILE: internal/handlers/auth_handler.go
// WHAT IT IS:     HTTP handlers for authentication endpoints
// WHY IT EXISTS:  Handles the HTTP layer for login/logout — reads
//                 the request, calls the service, sends the response
// ENDPOINTS:      POST /api/v1/auth/login
//                 POST /api/v1/auth/logout
// DEPENDS ON:     services/auth_service.go, utils/response.go
// IF REMOVED:     Admin cannot log in — dashboard is inaccessible
// LAST UPDATED:   2026-05-07 — initial creation
// ============================================================

package handlers

import (
	"imtiaz-portfolio/internal/services"
	"imtiaz-portfolio/internal/utils"

	"github.com/gin-gonic/gin"
)

// AuthHandler holds the auth service dependency.
type AuthHandler struct {
	authService *services.AuthService
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

/**
 * FUNCTION: Login
 * WHAT IT DOES:   Reads email + password from the JSON request body,
 *                 calls the auth service to verify credentials,
 *                 and returns a JWT token on success.
 * WHERE CALLED:   routes.go → POST /api/v1/auth/login
 * LAST UPDATED:   2026-05-07 — initial creation
 */
func (h *AuthHandler) Login(c *gin.Context) {
	// ─── Step 1: Parse and validate the request body ──────────
	var req services.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "Invalid request — email and password are required", err.Error())
		return
	}

	// ─── Step 2: Validate email format ────────────────────────
	if !utils.IsValidEmail(req.Email) {
		utils.BadRequest(c, "Please provide a valid email address", nil)
		return
	}

	// ─── Step 3: Attempt login via the service ────────────────
	response, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		// Return 401 for wrong credentials — don't reveal which field was wrong
		utils.Unauthorized(c, err.Error())
		return
	}

	// ─── Step 4: Return the token and user data ───────────────
	utils.Success(c, "Login successful", response)
}

/**
 * FUNCTION: Logout
 * WHAT IT DOES:   Logout endpoint — in a stateless JWT system,
 *                 logout is handled client-side by deleting the token
 *                 from localStorage. This endpoint just confirms success
 *                 so the frontend knows to redirect to the login page.
 * WHERE CALLED:   routes.go → POST /api/v1/auth/logout
 * LAST UPDATED:   2026-05-07 — initial creation
 */
func (h *AuthHandler) Logout(c *gin.Context) {
	// JWT tokens are stateless — there's no server-side session to invalidate.
	// The frontend must delete the token from localStorage on its own.
	// This endpoint just sends success so the client knows logout "worked".
	utils.Success(c, "Logged out successfully", nil)
}

/**
 * FUNCTION: Me
 * WHAT IT DOES:   Returns the currently authenticated user's info.
 *                 Used by the dashboard to display "Welcome, Imtiaz"
 *                 and to verify the stored token is still valid.
 * WHERE CALLED:   routes.go → GET /api/v1/auth/me [protected]
 * LAST UPDATED:   2026-05-07 — initial creation
 */
func (h *AuthHandler) Me(c *gin.Context) {
	// The auth middleware already validated the token and stored user info
	// in the context — we just read it out here
	userID, _ := c.Get("userID")
	email, _ := c.Get("email")
	role, _ := c.Get("role")

	utils.Success(c, "Authenticated user info", gin.H{
		"user_id": userID,
		"email":   email,
		"role":    role,
	})
}

func (h *AuthHandler) GetAccount(c *gin.Context) {
	userID := c.GetUint("userID")
	user, err := h.authService.GetByID(userID)
	if err != nil {
		utils.NotFound(c, err.Error())
		return
	}
	utils.Success(c, "Account fetched", user)
}

func (h *AuthHandler) UpdateName(c *gin.Context) {
	userID := c.GetUint("userID")
	var req struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Name == "" {
		utils.BadRequest(c, "Name is required", nil)
		return
	}
	if err := h.authService.UpdateName(userID, req.Name); err != nil {
		utils.InternalError(c, "Failed to update name")
		return
	}
	utils.Success(c, "Name updated successfully", nil)
}

func (h *AuthHandler) UpdateEmail(c *gin.Context) {
	userID := c.GetUint("userID")
	var req struct {
		Email           string `json:"email"`
		CurrentPassword string `json:"current_password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Email == "" || req.CurrentPassword == "" {
		utils.BadRequest(c, "Email and current password are required", nil)
		return
	}
	if err := h.authService.UpdateEmail(userID, req.Email, req.CurrentPassword); err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}
	utils.Success(c, "Email updated successfully", nil)
}

func (h *AuthHandler) UpdatePassword(c *gin.Context) {
	userID := c.GetUint("userID")
	var req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.CurrentPassword == "" || req.NewPassword == "" {
		utils.BadRequest(c, "Current and new password are required", nil)
		return
	}
	if len(req.NewPassword) < 6 {
		utils.BadRequest(c, "New password must be at least 6 characters", nil)
		return
	}
	if err := h.authService.UpdatePassword(userID, req.CurrentPassword, req.NewPassword); err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}
	utils.Success(c, "Password updated successfully", nil)
}
