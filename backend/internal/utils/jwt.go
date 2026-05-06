// ============================================================
// FILE: internal/utils/jwt.go
// WHAT IT IS:     JWT token generation and validation helpers
// WHY IT EXISTS:  Authentication requires creating signed tokens
//                 on login and verifying them on protected routes.
//                 Centralizing this prevents duplicate token logic.
// WHERE USED:     auth_service.go (generate), auth_middleware.go (verify)
// IF REMOVED:     Authentication completely breaks — no way to
//                 create or validate login tokens
// LAST UPDATED:   2026-05-07 — initial creation
// ============================================================

package utils

import (
	"errors"
	"time"

	// golang-jwt/jwt is the library that creates and verifies JWT tokens
	"github.com/golang-jwt/jwt/v5"
)

// Claims defines what information is embedded inside each JWT token.
// This data is readable by anyone who has the token (it's base64 encoded),
// but it cannot be tampered with because it's signed with the JWT_SECRET.
type Claims struct {
	// UserID identifies which user this token belongs to
	UserID uint `json:"user_id"`

	// Email included so the middleware can log which user made the request
	Email string `json:"email"`

	// Role is used for authorization checks (currently always "admin")
	Role string `json:"role"`

	// RegisteredClaims includes the standard JWT fields:
	// ExpiresAt, IssuedAt, Issuer, etc.
	jwt.RegisteredClaims
}

/**
 * FUNCTION: GenerateToken
 * WHAT IT DOES:   Creates a signed JWT token containing the user's
 *                 ID, email, and role. The token expires after the
 *                 duration specified in JWT_EXPIRY (e.g. "24h").
 *                 The frontend stores this token in localStorage
 *                 and sends it in the Authorization header on
 *                 every protected API request.
 * WHERE CALLED:   auth_service.go → Login() after password verified
 * PARAMETERS:     @param {uint} userID - the user's database ID
 *                 @param {string} email - the user's email address
 *                 @param {string} role - the user's role ("admin")
 *                 @param {string} secret - JWT_SECRET from config
 *                 @param {string} expiry - duration string e.g. "24h"
 * RETURNS:        @returns {string} - the signed JWT token string
 *                 @returns {error} - non-nil if token creation failed
 * EXAMPLE:        token, err := GenerateToken(1, "a@b.com", "admin", secret, "24h")
 * LAST UPDATED:   2026-05-07 — initial creation
 */
func GenerateToken(userID uint, email, role, secret, expiry string) (string, error) {
	// ─── Step 1: Parse the expiry duration ────────────────────
	// time.ParseDuration converts "24h" → 24 hours as time.Duration
	duration, err := time.ParseDuration(expiry)
	if err != nil {
		return "", errors.New("invalid JWT_EXPIRY format — use values like '24h', '30m'")
	}

	// ─── Step 2: Build the claims payload ─────────────────────
	// ExpiresAt is now + duration — after this time the token is invalid
	claims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "imtiaz-portfolio",
		},
	}

	// ─── Step 3: Sign the token with HMAC-SHA256 ──────────────
	// HS256 uses the JWT_SECRET to sign the claims.
	// Anyone with the secret can verify the token is authentic.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

/**
 * FUNCTION: ValidateToken
 * WHAT IT DOES:   Parses and validates a JWT token string.
 *                 Checks that: the signature is valid (not tampered),
 *                 the token has not expired, and the algorithm matches.
 *                 Returns the Claims so the middleware can access UserID.
 * WHERE CALLED:   auth_middleware.go — on every protected API request
 * PARAMETERS:     @param {string} tokenString - the raw JWT from Authorization header
 *                 @param {string} secret - JWT_SECRET from config (must match)
 * RETURNS:        @returns {*Claims} - the decoded claims if valid
 *                 @returns {error} - non-nil if token is invalid or expired
 * EXAMPLE:        claims, err := ValidateToken(tokenStr, cfg.JWTSecret)
 * LAST UPDATED:   2026-05-07 — initial creation
 */
func ValidateToken(tokenString, secret string) (*Claims, error) {
	// jwt.ParseWithClaims decodes the token and verifies the signature.
	// The callback function is called to supply the signing key — this
	// is where we enforce that only HS256 tokens are accepted (preventing
	// the "alg:none" attack where someone strips the signature)
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Reject any token that wasn't signed with HMAC (e.g. RSA, none)
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method — token may be tampered")
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	// Extract the claims from the verified token
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}
