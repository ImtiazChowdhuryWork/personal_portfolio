// ============================================================
// FILE: internal/services/auth_service.go
// WHAT IT IS:     Authentication business logic
// WHY IT EXISTS:  Handles login validation, password checking,
//                 and token generation. Keeps auth logic out of
//                 handlers so it can be tested independently.
// ENDPOINTS:      Called by auth_handler.go
// DEPENDS ON:     models/user.go, utils/jwt.go, utils/hash.go, config
// IF REMOVED:     Login endpoint has no logic — authentication fails
// LAST UPDATED:   2026-05-07 — initial creation
// ============================================================

package services

import (
	"errors"
	"imtiaz-portfolio/config"
	"imtiaz-portfolio/internal/models"
	"imtiaz-portfolio/internal/utils"

	"gorm.io/gorm"
)

// AuthService holds the dependencies needed for authentication operations.
type AuthService struct {
	db  *gorm.DB        // database connection for looking up users
	cfg *config.Config  // config for JWT secret and expiry settings
}

/**
 * FUNCTION: NewAuthService
 * WHAT IT DOES:   Creates a new AuthService with the given dependencies.
 *                 This pattern (constructor injection) makes the service
 *                 testable by allowing fake DB/config to be passed in.
 * WHERE CALLED:   cmd/main.go at startup
 * LAST UPDATED:   2026-05-07 — initial creation
 */
func NewAuthService(db *gorm.DB, cfg *config.Config) *AuthService {
	return &AuthService{db: db, cfg: cfg}
}

// LoginRequest is the expected JSON body for the login endpoint.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginResponse is returned on successful login.
type LoginResponse struct {
	Token   string      `json:"token"`
	User    models.User `json:"user"`
}

/**
 * FUNCTION: Login
 * WHAT IT DOES:   1. Finds the user by email in the database
 *                 2. Verifies the password matches the stored bcrypt hash
 *                 3. Generates a JWT token if everything is valid
 *                 4. Returns the token and user info to the handler
 * WHERE CALLED:   auth_handler.go → POST /api/v1/auth/login
 * PARAMETERS:     @param {string} email - the email the user typed
 *                 @param {string} password - the plain text password typed
 * RETURNS:        @returns {*LoginResponse} - token + user data on success
 *                 @returns {error} - description of what went wrong
 * LAST UPDATED:   2026-05-07 — initial creation
 */
func (s *AuthService) Login(email, password string) (*LoginResponse, error) {
	// ─── Step 1: Find the user by email ───────────────────────
	var user models.User
	result := s.db.Where("email = ?", email).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// Return a generic message — don't tell attackers which field was wrong
			return nil, errors.New("invalid email or password")
		}
		return nil, errors.New("database error during login")
	}

	// ─── Step 2: Verify the password ──────────────────────────
	// CheckPassword runs bcrypt comparison — slow by design to defeat brute force
	if !utils.CheckPassword(password, user.Password) {
		return nil, errors.New("invalid email or password")
	}

	// ─── Step 3: Generate the JWT access token ────────────────
	token, err := utils.GenerateToken(user.ID, user.Email, user.Role, s.cfg.JWTSecret, s.cfg.JWTExpiry)
	if err != nil {
		return nil, errors.New("failed to generate authentication token")
	}

	return &LoginResponse{
		Token: token,
		User:  user,
	}, nil
}

func (s *AuthService) GetByID(id uint) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, id).Error; err != nil {
		return nil, errors.New("user not found")
	}
	return &user, nil
}

func (s *AuthService) UpdateName(userID uint, name string) error {
	return s.db.Model(&models.User{}).Where("id = ?", userID).Update("name", name).Error
}

func (s *AuthService) UpdateEmail(userID uint, email, currentPassword string) error {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return errors.New("user not found")
	}
	if !utils.CheckPassword(currentPassword, user.Password) {
		return errors.New("current password is incorrect")
	}
	var count int64
	s.db.Model(&models.User{}).Where("email = ? AND id != ?", email, userID).Count(&count)
	if count > 0 {
		return errors.New("email is already in use")
	}
	return s.db.Model(&user).Update("email", email).Error
}

func (s *AuthService) UpdatePassword(userID uint, currentPassword, newPassword string) error {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return errors.New("user not found")
	}
	if !utils.CheckPassword(currentPassword, user.Password) {
		return errors.New("current password is incorrect")
	}
	hashed, err := utils.HashPassword(newPassword)
	if err != nil {
		return errors.New("failed to hash new password")
	}
	return s.db.Model(&user).Update("password", hashed).Error
}
