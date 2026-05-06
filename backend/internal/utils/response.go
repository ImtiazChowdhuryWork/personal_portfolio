// ============================================================
// FILE: internal/utils/response.go
// WHAT IT IS:     Standardized API response helpers
// WHY IT EXISTS:  Every API endpoint must return the same JSON
//                 shape so the frontend can handle all responses
//                 with one piece of code instead of checking
//                 different formats per endpoint
// WHERE USED:     Every handler file calls these functions
// IF REMOVED:     Each handler would format responses differently —
//                 the frontend would break unpredictably
// LAST UPDATED:   2026-05-07 — initial creation
// ============================================================

package utils

import (
	"net/http"

	// Gin is the HTTP framework — c.JSON writes a JSON response
	"github.com/gin-gonic/gin"
)

// APIResponse is the standard shape for EVERY response from this API.
// The frontend's api.js always expects this exact structure.
//
// Success response example:
//
//	{ "success": true, "message": "Project created", "data": {...}, "error": null }
//
// Error response example:
//
//	{ "success": false, "message": "Validation failed", "data": null, "error": "name is required" }
type APIResponse struct {
	Success bool        `json:"success"`          // true if the request succeeded
	Message string      `json:"message"`          // human-readable status description
	Data    interface{} `json:"data"`             // the actual response payload (or null)
	Error   interface{} `json:"error"`            // error details (or null on success)
}

/**
 * FUNCTION: Success
 * WHAT IT DOES:   Sends a 200 OK JSON response with success=true,
 *                 the provided message, and the data payload
 * WHERE CALLED:   All handler functions after successful operations
 * PARAMETERS:     @param {*gin.Context} c - the current HTTP request context
 *                 @param {string} message - description of what happened
 *                 @param {interface{}} data - the data to return to the client
 * EXAMPLE:
 *   utils.Success(c, "Projects fetched", projects)
 * LAST UPDATED:   2026-05-07 — initial creation
 */
func Success(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
		Error:   nil,
	})
}

/**
 * FUNCTION: Created
 * WHAT IT DOES:   Sends a 201 Created JSON response — used when
 *                 a new resource (project, skill, etc.) was just created
 * WHERE CALLED:   Handler POST routes after successful creation
 * PARAMETERS:     @param {*gin.Context} c - HTTP context
 *                 @param {string} message - creation success message
 *                 @param {interface{}} data - the newly created object
 * LAST UPDATED:   2026-05-07 — initial creation
 */
func Created(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusCreated, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
		Error:   nil,
	})
}

/**
 * FUNCTION: BadRequest
 * WHAT IT DOES:   Sends a 400 Bad Request response when the client
 *                 sent invalid data (missing fields, wrong format)
 * WHERE CALLED:   Handlers when JSON binding or validation fails
 * PARAMETERS:     @param {*gin.Context} c - HTTP context
 *                 @param {string} message - what was wrong with the request
 *                 @param {interface{}} err - the specific validation error(s)
 * LAST UPDATED:   2026-05-07 — initial creation
 */
func BadRequest(c *gin.Context, message string, err interface{}) {
	c.JSON(http.StatusBadRequest, APIResponse{
		Success: false,
		Message: message,
		Data:    nil,
		Error:   err,
	})
}

/**
 * FUNCTION: Unauthorized
 * WHAT IT DOES:   Sends a 401 Unauthorized response when the request
 *                 has no valid JWT token or the token has expired
 * WHERE CALLED:   auth_middleware.go when token is missing or invalid
 * PARAMETERS:     @param {*gin.Context} c - HTTP context
 *                 @param {string} message - reason for unauthorized
 * LAST UPDATED:   2026-05-07 — initial creation
 */
func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, APIResponse{
		Success: false,
		Message: message,
		Data:    nil,
		Error:   "authentication required",
	})
}

/**
 * FUNCTION: NotFound
 * WHAT IT DOES:   Sends a 404 Not Found response when the requested
 *                 resource does not exist in the database
 * WHERE CALLED:   Handlers when a DB lookup returns "record not found"
 * PARAMETERS:     @param {*gin.Context} c - HTTP context
 *                 @param {string} message - what was not found
 * LAST UPDATED:   2026-05-07 — initial creation
 */
func NotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, APIResponse{
		Success: false,
		Message: message,
		Data:    nil,
		Error:   "resource not found",
	})
}

/**
 * FUNCTION: InternalError
 * WHAT IT DOES:   Sends a 500 Internal Server Error response for
 *                 unexpected failures (database down, disk full, etc.)
 *                 Does NOT expose internal error details to the client
 *                 in production — logs them server-side instead
 * WHERE CALLED:   Handlers when an unexpected error occurs
 * PARAMETERS:     @param {*gin.Context} c - HTTP context
 *                 @param {string} message - safe error description for client
 * LAST UPDATED:   2026-05-07 — initial creation
 */
func InternalError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, APIResponse{
		Success: false,
		Message: message,
		Data:    nil,
		Error:   "internal server error",
	})
}
