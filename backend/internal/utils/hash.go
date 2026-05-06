// ============================================================
// FILE: internal/utils/hash.go
// WHAT IT IS:     Password hashing and verification helpers
// WHY IT EXISTS:  Passwords must NEVER be stored in plain text.
//                 bcrypt turns "Admin@1234" into a long random-looking
//                 string that cannot be reversed — even if the database
//                 is stolen, passwords remain safe.
// WHERE USED:     auth_service.go (hash on register/seed),
//                 auth_service.go (verify on login)
// IF REMOVED:     Passwords would be stored as plain text — severe security risk
// LAST UPDATED:   2026-05-07 — initial creation
// ============================================================

package utils

import (
	// bcrypt is the industry standard for password hashing
	// It is deliberately slow to make brute-force attacks impractical
	"golang.org/x/crypto/bcrypt"
)

/**
 * FUNCTION: HashPassword
 * WHAT IT DOES:   Takes a plain text password (e.g. "Admin@1234")
 *                 and returns a bcrypt hash (e.g. "$2a$12$...")
 *                 that is safe to store in the database.
 *                 Cost factor 12 means ~300ms per hash — fast enough
 *                 for login, slow enough to defeat brute-force attacks.
 * WHERE CALLED:   seeds/seed.go when creating the default admin,
 *                 auth_handler.go if a register endpoint is added
 * PARAMETERS:     @param {string} password - the raw password to hash
 * RETURNS:        @returns {string} - the bcrypt hash string
 *                 @returns {error} - non-nil if hashing failed
 * EXAMPLE:        hash, err := HashPassword("Admin@1234")
 * IF CHANGED:     Cost factor change requires re-hashing all passwords
 * LAST UPDATED:   2026-05-07 — initial creation
 */
func HashPassword(password string) (string, error) {
	// Cost 12 is a good balance: ~300ms to hash on modern hardware.
	// Higher = slower and more secure but impacts login UX.
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(bytes), err
}

/**
 * FUNCTION: CheckPassword
 * WHAT IT DOES:   Compares a plain text password against a stored
 *                 bcrypt hash and returns true if they match.
 *                 bcrypt internally re-runs the same hash algorithm
 *                 on the plain password and compares the result.
 * WHERE CALLED:   auth_service.go → Login() to verify login attempts
 * PARAMETERS:     @param {string} password - the raw password the user typed
 *                 @param {string} hash - the bcrypt hash from the database
 * RETURNS:        @returns {bool} - true if password matches the hash
 * EXAMPLE:        ok := CheckPassword("Admin@1234", user.Password)
 * LAST UPDATED:   2026-05-07 — initial creation
 */
func CheckPassword(password, hash string) bool {
	// bcrypt.CompareHashAndPassword returns nil if they match,
	// or an error if they don't — we convert that to a simple bool
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
