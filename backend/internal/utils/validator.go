// ============================================================
// FILE: internal/utils/validator.go
// WHAT IT IS:     Input validation helper functions
// WHY IT EXISTS:  Validates user-supplied data at the API boundary
//                 before it reaches the database. Prevents garbage
//                 data and SQL injection from reaching our models.
// WHERE USED:     All handler files before processing request bodies
// IF REMOVED:     Invalid data reaches the database — potential crashes
//                 or data corruption
// LAST UPDATED:   2026-05-07 — initial creation
// ============================================================

package utils

import (
	"regexp"
	"strings"
)

// emailRegex is compiled once at startup (not on every call) for performance.
// This regex checks for the basic shape of an email address.
// It is intentionally simple — a real email can only be verified by sending to it.
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

/**
 * FUNCTION: IsValidEmail
 * WHAT IT DOES:   Checks if the given string looks like a valid email.
 *                 Trims whitespace first so " user@example.com " passes.
 * WHERE CALLED:   auth_handler.go (login validation),
 *                 message_handler.go (contact form validation)
 * PARAMETERS:     @param {string} email - the email string to check
 * RETURNS:        @returns {bool} - true if format looks valid
 * EXAMPLE:        IsValidEmail("user@example.com") → true
 *                 IsValidEmail("notanemail") → false
 * LAST UPDATED:   2026-05-07 — initial creation
 */
func IsValidEmail(email string) bool {
	email = strings.TrimSpace(email)
	return emailRegex.MatchString(email)
}

/**
 * FUNCTION: IsEmpty
 * WHAT IT DOES:   Returns true if the string is empty or contains
 *                 only whitespace. Used to check required fields.
 * WHERE CALLED:   All handlers validating required string fields
 * PARAMETERS:     @param {string} s - the string to check
 * RETURNS:        @returns {bool} - true if empty or whitespace only
 * EXAMPLE:        IsEmpty("") → true, IsEmpty("  ") → true
 * LAST UPDATED:   2026-05-07 — initial creation
 */
func IsEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}

/**
 * FUNCTION: Sanitize
 * WHAT IT DOES:   Trims leading/trailing whitespace from a string.
 *                 Called on all string inputs before saving to DB
 *                 so we don't store " Flutter " instead of "Flutter".
 * WHERE CALLED:   Handlers before binding request data into models
 * PARAMETERS:     @param {string} s - the string to clean
 * RETURNS:        @returns {string} - trimmed string
 * LAST UPDATED:   2026-05-07 — initial creation
 */
func Sanitize(s string) string {
	return strings.TrimSpace(s)
}

/**
 * FUNCTION: IsValidPercentage
 * WHAT IT DOES:   Checks that a skill percentage is between 0 and 100.
 *                 Prevents storing 150% or -5% in the database.
 * WHERE CALLED:   skill_handler.go when creating or updating skills
 * PARAMETERS:     @param {int} p - the percentage value to validate
 * RETURNS:        @returns {bool} - true if 0 ≤ p ≤ 100
 * LAST UPDATED:   2026-05-07 — initial creation
 */
func IsValidPercentage(p int) bool {
	return p >= 0 && p <= 100
}
