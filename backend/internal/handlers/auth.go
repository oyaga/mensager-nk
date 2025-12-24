package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/nakamura/chatwoot-go/internal/config"
	"github.com/nakamura/chatwoot-go/internal/middleware"
	"github.com/nakamura/chatwoot-go/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthHandler struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewAuthHandler(db *gorm.DB, cfg *config.Config) *AuthHandler {
	return &AuthHandler{db: db, cfg: cfg}
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginResponse struct {
	Token string       `json:"token"`
	User  *models.User `json:"user"`
}

// Login authenticates a user
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Login error: Failed to bind JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Login attempt for email: %s", req.Email)

	// Find user by email
	var user models.User
	if err := h.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		log.Printf("Login error: User not found for email %s: %v", req.Email, err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	log.Printf("User found: %s, hash length: %d", user.Email, len(user.PasswordHash))

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		log.Printf("Login error: Password mismatch for %s: %v", req.Email, err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	log.Printf("Password verified for user: %s", user.Email)

	// Get user's first account (for simplicity)
	var accountUser models.AccountUser
	if err := h.db.Where("user_id = ?", user.ID).First(&accountUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User has no associated account"})
		return
	}

	// Generate JWT token
	token, err := h.generateToken(&user, accountUser.AccountID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Hide password hash
	user.PasswordHash = ""

	c.JSON(http.StatusOK, LoginResponse{
		Token: token,
		User:  &user,
	})
}

// Register creates a new user account
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user already exists
	var existingUser models.User
	if err := h.db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Create user and account in a transaction
	var user models.User
	var account models.Account

	err = h.db.Transaction(func(tx *gorm.DB) error {
		// Create account
		account = models.Account{
			Name:   req.Name + "'s Account",
			Status: "active",
			Locale: "en",
		}
		if err := tx.Create(&account).Error; err != nil {
			return err
		}

		// Create user
		user = models.User{
			Name:         req.Name,
			Email:        req.Email,
			PasswordHash: string(hashedPassword),
			DisplayName:  req.Name,
			Role:         "administrator",
			Availability: "online",
		}
		if err := tx.Create(&user).Error; err != nil {
			return err
		}

		// Associate user with account
		accountUser := models.AccountUser{
			AccountID: account.ID,
			UserID:    user.ID,
			Role:      "administrator",
		}
		if err := tx.Create(&accountUser).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Generate JWT token
	token, err := h.generateToken(&user, account.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Hide password hash
	user.PasswordHash = ""

	c.JSON(http.StatusCreated, LoginResponse{
		Token: token,
		User:  &user,
	})
}

// GetProfile returns the current user's profile
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var user models.User
	if err := h.db.Preload("Accounts").First(&user, userID).Error; err != nil {
		log.Printf("GetProfile - Error finding user: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Ensure AccessToken exists (migration for existing users)
	if user.AccessToken == "" {
		user.AccessToken = uuid.New().String()
		h.db.Model(&user).Update("access_token", user.AccessToken)
	}

	// DEBUG: Log loaded accounts
	log.Printf("GetProfile - User %s, Loaded Accounts via Preload: %d", user.Email, len(user.Accounts))

	// Force load accounts if empty (fallback)
	if len(user.Accounts) == 0 {
		var accounts []models.Account
		err := h.db.Model(&user).Association("Accounts").Find(&accounts)
		if err != nil {
			log.Printf("GetProfile - Error manual loading accounts: %v", err)
		} else {
			user.Accounts = accounts
			log.Printf("GetProfile - Manually loaded Accounts: %d", len(user.Accounts))
		}
	}

	// ULTIMATE FALLBACK: Raw SQL
	if len(user.Accounts) == 0 {
		var accountID uuid.UUID
		err := h.db.Raw("SELECT account_id FROM account_users WHERE user_id = ?", user.ID).Scan(&accountID).Error
		if err == nil && accountID != uuid.Nil {
			log.Printf("GetProfile - Raw SQL found account: %s", accountID)
			user.Accounts = []models.Account{{
				BaseModel: models.BaseModel{ID: accountID},
				Name:      "Account (Recovered)",
			}}
		} else if err != nil {
			log.Printf("GetProfile - Raw SQL error: %v", err)
		}
	}

	user.PasswordHash = ""
	c.JSON(http.StatusOK, user)
}

// UpdateProfile updates the current user's profile
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req struct {
		Name        string         `json:"name"`
		Email       string         `json:"email"`
		DisplayName string         `json:"display_name"`
		Avatar      string         `json:"avatar"`
		UISettings  map[string]any `json:"ui_settings"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.DisplayName != "" {
		updates["display_name"] = req.DisplayName
	}
	if req.Avatar != "" {
		updates["avatar"] = req.Avatar
	}
	if req.UISettings != nil {
		updates["ui_settings"] = req.UISettings
	}

	// Handle Email update separately to check for duplicates
	if req.Email != "" {
		var existingUser models.User
		// Check if any OTHER user has this email
		if err := h.db.Where("email = ? AND id != ?", req.Email, userID).First(&existingUser).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already in use"})
			return
		}
		updates["email"] = req.Email
	}

	if err := h.db.Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	// Fetch updated user
	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch updated user"})
		return
	}
	user.PasswordHash = ""

	// Generate new token with updated user data
	accountID := c.MustGet("account_id").(uuid.UUID)
	newToken, err := h.generateToken(&user, accountID)
	if err != nil {
		// Log error but return user at least
		// If token generation fails, return the user without a token
		c.JSON(http.StatusOK, gin.H{
			"user": user,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":  user,
		"token": newToken,
	})
}

// ResetAccessToken generates a new access token for the user
func (h *AuthHandler) ResetAccessToken(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	newToken := uuid.New().String()

	if err := h.db.Model(&models.User{}).Where("id = ?", userID).Update("access_token", newToken).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"access_token": newToken})
}

// ChangePassword changes the user's password
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Current password is incorrect"})
		return
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Update password
	if err := h.db.Model(&user).Update("password_hash", string(hashedPassword)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}

// UpdateAvailability updates user's availability status
func (h *AuthHandler) UpdateAvailability(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req struct {
		Availability string `json:"availability" binding:"required,oneof=online busy offline"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.db.Model(&models.User{}).Where("id = ?", userID).Update("availability", req.Availability).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update availability"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"availability": req.Availability})
}

// ForgotPassword initiates password reset
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	// TODO: Implement email sending
	c.JSON(http.StatusOK, gin.H{"message": "Password reset email sent"})
}

// ResetPassword resets user password
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	// TODO: Implement password reset with token
	c.JSON(http.StatusOK, gin.H{"message": "Password reset successful"})
}

// ListUsers lists all users (admin only)
func (h *AuthHandler) ListUsers(c *gin.Context) {
	var users []models.User
	if err := h.db.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	// Hide password hashes
	for i := range users {
		users[i].PasswordHash = ""
	}

	c.JSON(http.StatusOK, users)
}

// UpdateUserRole updates a user's role (admin only)
func (h *AuthHandler) UpdateUserRole(c *gin.Context) {
	userID := c.Param("id")

	var req struct {
		Role string `json:"role" binding:"required,oneof=administrator agent supervisor"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.db.Model(&models.User{}).Where("id = ?", userID).Update("role", req.Role).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role updated successfully"})
}

// generateToken generates a JWT token for a user
func (h *AuthHandler) generateToken(user *models.User, accountID uuid.UUID) (string, error) {
	claims := middleware.Claims{
		UserID:    user.ID,
		Email:     user.Email,
		Role:      user.Role,
		AccountID: accountID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.cfg.JWTSecret))
}
