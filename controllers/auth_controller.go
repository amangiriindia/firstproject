package controllers

import (
	"firstproject/database"
	"firstproject/middleware"
	"firstproject/models"
	"firstproject/utils"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"strings"
	"time"
)

// Register creates a new user account
func Register(c *fiber.Ctx) error {
	// Custom input struct for registration
	type RegisterInput struct {
		Email     string `json:"email" validate:"required,email"`
		Password  string `json:"password" validate:"required,min=6"`
		Username  string `json:"username" validate:"required,min=3,max=50"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Role      string `json:"role" validate:"required,oneof=user instructor admin super_admin"`
	}

	var input RegisterInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	// Validate input
	if input.Email == "" || input.Password == "" || input.Username == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Email, password, and username are required"})
	}

	// Validate role
	validRoles := map[string]bool{
		"user":        true,
		"instructor":  true,
		"admin":       true,
		"super_admin": true,
	}

	if input.Role == "" {
		input.Role = "user" // Default role
	} else if !validRoles[strings.ToLower(input.Role)] {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid role. Valid roles are: user, instructor, admin, super_admin",
		})
	}

	// Check if user already exists (email or username)
	var existingUser models.User
	if err := database.DB.Where("email = ? OR username = ?",
		strings.ToLower(input.Email),
		strings.ToLower(input.Username)).First(&existingUser).Error; err == nil {
		if existingUser.Email == strings.ToLower(input.Email) {
			return c.Status(400).JSON(fiber.Map{"error": "Email already in use"})
		}
		if existingUser.Username == strings.ToLower(input.Username) {
			return c.Status(400).JSON(fiber.Map{"error": "Username already in use"})
		}
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), 14)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not hash password"})
	}

	// Create user
	user := models.User{
		Username:          strings.ToLower(input.Username),
		Email:             strings.ToLower(input.Email),
		PasswordHash:      string(hash),
		FirstName:         input.FirstName,
		LastName:          input.LastName,
		Role:              strings.ToLower(input.Role),
		VerificationToken: utils.GenerateRandomString(32),
	}

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
	token, err := middleware.GenerateJWT(user)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not generate token"})
	}

	return c.JSON(fiber.Map{
		"token": token,
		"user":  user,
	})
}

// Rest of your controller functions remain the same...
func Login(c *fiber.Ctx) error {
	reqData := new(struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	})

	// Body parsing error handling
	if err := c.BodyParser(reqData); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	// Log the input for debugging (optional)
	fmt.Println("Input Email:", reqData.Email)
	fmt.Println("Input Password:", reqData.Password)

	// Find user by email
	var user models.User
	if err := database.DB.Where("email = ?", strings.ToLower(reqData.Email)).First(&user).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	// Compare password with the hashed password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(reqData.Password)); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	// Generate JWT token
	token, err := middleware.GenerateJWT(user)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not generate token"})
	}

	// Return success response with token and user details
	return c.JSON(fiber.Map{
		"token": token,
		"user":  user,
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

	fmt.Println(" Reset token:", resetToken)
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
	fmt.Print(user.VerificationToken)
	// Send verification email
	go utils.SendVerificationEmail(user.Email, user.VerificationToken)

	return c.JSON(fiber.Map{
		"message": "If an unverified account with that email exists, a new verification email has been sent",
	})
}
