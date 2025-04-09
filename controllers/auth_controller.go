package controllers

import (
	"firstproject/database"
	"firstproject/models"
	"firstproject/utils"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Register creates a new user account
func Register(c *fiber.Ctx) error {
	user := new(models.User)
	if err := c.BodyParser(user); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	// Validate input
	if user.Email == "" || user.Password == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Email and password are required"})
	}

	// Check if user already exists
	var existingUser models.User
	database.DB.Where("email = ?", strings.ToLower(user.Email)).First(&existingUser)
	if existingUser.ID != 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Email already in use"})
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), 14)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not hash password"})
	}
	user.PasswordHash = string(hash) // <-- save to PasswordHash

	user.Password = string(hash)
	user.Email = strings.ToLower(user.Email)
	user.VerificationToken = utils.GenerateRandomString(32)

	// Create user
	if err := database.DB.Create(&user).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not create user"})
	}

	// Send verification email
	go utils.SendVerificationEmail(user.Email, user.VerificationToken)

	// Create profile
	profile := models.UserProfile{
		UserID: user.ID,
	}
	database.DB.Create(&profile)

	// Generate JWT
	token, err := generateJWT(*user) // Pass the struct value instead of pointer
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not generate token"})
	}

	return c.JSON(fiber.Map{
		"token": token,
		"user":  user,
	})
}

// Login authenticates a user
func Login(c *fiber.Ctx) error {
	input := new(models.LoginInput)
	if err := c.BodyParser(input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	fmt.Println("Input Email:", input.Email)
	fmt.Println("Input Password:", input.Password)
	var user models.User
	database.DB.Where("email = ?", strings.ToLower(input.Email)).First(&user)
	if user.ID == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	// Generate JWT
	token, err := generateJWT(user)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not generate token"})
	}

	// Store session
	session := models.UserSession{
		UserID:    user.ID,
		Token:     token,
		IPAddress: c.IP(),
		UserAgent: c.Get("User-Agent"),
		ExpiresAt: time.Now().Add(time.Hour * 72),
	}
	database.DB.Create(&session)

	return c.JSON(fiber.Map{
		"token": token,
		"user":  user,
	})
}

// RefreshToken generates a new access token
func RefreshToken(c *fiber.Ctx) error {
	refreshToken := c.Get("Authorization")
	if refreshToken == "" {
		return c.Status(401).JSON(fiber.Map{"error": "Refresh token missing"})
	}

	// Remove "Bearer " prefix if present
	refreshToken = strings.Replace(refreshToken, "Bearer ", "", 1)

	// Parse and validate the refresh token
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_REFRESH_SECRET")), nil
	})
	if err != nil || !token.Valid {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid refresh token"})
	}

	claims := token.Claims.(jwt.MapClaims)
	userID := uint(claims["user_id"].(float64))

	// Check if refresh token exists in database
	var session models.UserSession
	if err := database.DB.Where("token = ? AND expires_at > ?", refreshToken, time.Now()).First(&session).Error; err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid refresh token"})
	}

	// Get user
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "User not found"})
	}

	// Generate new access token
	newToken, err := generateJWT(user)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not generate token"})
	}

	return c.JSON(fiber.Map{
		"token": newToken,
	})
}

// Logout invalidates the current session
func Logout(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID := uint(claims["user_id"].(float64))

	// Get token from header
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(401).JSON(fiber.Map{"error": "Authorization header missing"})
	}
	tokenString := strings.Replace(authHeader, "Bearer ", "", 1)

	// Delete the session
	if err := database.DB.Where("user_id = ? AND token = ?", userID, tokenString).Delete(&models.UserSession{}).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not logout"})
	}

	return c.JSON(fiber.Map{
		"message": "Successfully logged out",
	})
}

// ForgotPassword initiates password reset
func ForgotPassword(c *fiber.Ctx) error {
	var input struct {
		Email string `json:"email"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	// Find user by email
	var user models.User
	if err := database.DB.Where("email = ?", strings.ToLower(input.Email)).First(&user).Error; err != nil {
		// Explicitly notify user not found
		return c.Status(404).JSON(fiber.Map{
			"error":   "User does not exist with this email address",
			"success": false,
		})
	}

	// Generate reset token
	resetToken := utils.GenerateRandomString(32)
	user.ResetToken = resetToken
	user.ResetTokenExpires = time.Now().Add(time.Hour * 1) // Token expires in 1 hour

	if err := database.DB.Save(&user).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not generate reset token"})
	}

	// Send email with reset link (asynchronously)
	go utils.SendPasswordResetEmail(user.Email, resetToken)

	return c.JSON(fiber.Map{
		"message": fmt.Sprintf("A password reset link has been sent to %s", user.Email),
		"success": true,
	})
}

// ResetPassword handles password reset
func ResetPassword(c *fiber.Ctx) error {
	var input struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	// Find user by reset token
	var user models.User
	if err := database.DB.Where("reset_token = ? AND reset_token_expires > ?", input.Token, time.Now()).First(&user).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid or expired reset token"})
	}

	// Hash new password
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), 14)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not hash password"})
	}

	// Update password and clear reset token
	user.Password = string(hash)
	user.ResetToken = ""
	user.ResetTokenExpires = time.Time{}

	if err := database.DB.Save(&user).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not update password"})
	}

	// Invalidate all existing sessions
	database.DB.Where("user_id = ?", user.ID).Delete(&models.UserSession{})

	return c.JSON(fiber.Map{
		"message": "Password successfully reset",
	})
}

// VerifyEmail confirms user's email address
func VerifyEmail(c *fiber.Ctx) error {
	var input struct {
		Token string `json:"token"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	// Find user by verification token
	var user models.User
	if err := database.DB.Where("verification_token = ?", input.Token).First(&user).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid verification token"})
	}

	// Mark as verified
	user.IsVerified = true
	user.VerificationToken = ""

	if err := database.DB.Save(&user).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not verify email"})
	}

	return c.JSON(fiber.Map{
		"message": "Email successfully verified",
	})
}

// ResendVerification sends a new verification email
func ResendVerification(c *fiber.Ctx) error {
	var input struct {
		Email string `json:"email"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	// Find user by email
	var user models.User
	if err := database.DB.Where("email = ? AND is_verified = ?", strings.ToLower(input.Email), false).First(&user).Error; err != nil {
		// Don't reveal if user doesn't exist or is already verified
		return c.JSON(fiber.Map{
			"message": "If an unverified account with that email exists, a new verification email has been sent",
		})
	}

	// Generate new verification token if none exists
	if user.VerificationToken == "" {
		user.VerificationToken = utils.GenerateRandomString(32)
		if err := database.DB.Save(&user).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Could not generate verification token"})
		}
	}

	// Send verification email
	go utils.SendVerificationEmail(user.Email, user.VerificationToken)

	return c.JSON(fiber.Map{
		"message": "If an unverified account with that email exists, a new verification email has been sent",
	})
}

// GetCurrentUser returns the authenticated user's profile
func GetCurrentUser(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID := uint(claims["user_id"].(float64))

	var dbUser models.User
	if err := database.DB.Preload("Profile").First(&dbUser, userID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "User not found"})
	}

	// Don't return sensitive fields
	dbUser.Password = ""
	dbUser.VerificationToken = ""
	dbUser.ResetToken = ""
	dbUser.ResetTokenExpires = time.Time{}

	return c.JSON(dbUser)
}

// UpdateProfile updates the authenticated user's profile
func UpdateProfile(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID := uint(claims["user_id"].(float64))

	// Get the existing user
	var dbUser models.User
	if err := database.DB.Preload("Profile").First(&dbUser, userID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "User not found"})
	}

	// Parse input
	var input struct {
		FirstName   *string  `json:"first_name"`
		LastName    *string  `json:"last_name"`
		AvatarURL   *string  `json:"avatar_url"`
		Bio         *string  `json:"bio"`
		Skills      []string `json:"skills"`
		Interests   []string `json:"interests"`
		GithubURL   *string  `json:"github_url"`
		LinkedinURL *string  `json:"linkedin_url"`
		TwitterURL  *string  `json:"twitter_url"`
		WebsiteURL  *string  `json:"website_url"`
		Education   *string  `json:"education"`
		Experience  *string  `json:"experience"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	// Update user fields if provided
	if input.FirstName != nil {
		dbUser.FirstName = *input.FirstName
	}
	if input.LastName != nil {
		dbUser.LastName = *input.LastName
	}
	if input.AvatarURL != nil {
		dbUser.AvatarURL = *input.AvatarURL
	}
	if input.Bio != nil {
		dbUser.Bio = *input.Bio
	}

	// Update profile fields if provided
	if input.Skills != nil {
		dbUser.Profile.Skills = input.Skills
	}
	if input.Interests != nil {
		dbUser.Profile.Interests = input.Interests
	}
	if input.GithubURL != nil {
		dbUser.Profile.GithubURL = *input.GithubURL
	}
	if input.LinkedinURL != nil {
		dbUser.Profile.LinkedinURL = *input.LinkedinURL
	}
	if input.TwitterURL != nil {
		dbUser.Profile.TwitterURL = *input.TwitterURL
	}
	if input.WebsiteURL != nil {
		dbUser.Profile.WebsiteURL = *input.WebsiteURL
	}
	if input.Education != nil {
		dbUser.Profile.Education = *input.Education
	}
	if input.Experience != nil {
		dbUser.Profile.Experience = *input.Experience
	}

	// Save changes
	if err := database.DB.Save(&dbUser).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not update user"})
	}
	if err := database.DB.Save(&dbUser.Profile).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not update profile"})
	}

	// Don't return sensitive fields
	dbUser.Password = ""
	dbUser.VerificationToken = ""
	dbUser.ResetToken = ""
	dbUser.ResetTokenExpires = time.Time{}

	return c.JSON(dbUser)
}

// Helper function to generate JWT
func generateJWT(user models.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}
