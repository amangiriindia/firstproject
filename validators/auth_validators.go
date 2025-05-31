package validators

import (
	"github.com/gofiber/fiber/v2"
	"strings"
)

// RegisterValidator validates the input for the registration route
func RegisterValidator(c *fiber.Ctx) error {
	var input struct {
		Email     string `json:"email"`
		Password  string `json:"password"`
		Username  string `json:"username"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Role      string `json:"role"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	// Validate required fields
	if input.Email == "" || input.Password == "" || input.Username == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Email, password, and username are required"})
	}

	// Validate password length
	if len(input.Password) < 6 {
		return c.Status(400).JSON(fiber.Map{"error": "Password must be at least 6 characters long"})
	}

	// Validate username length
	if len(input.Username) < 3 || len(input.Username) > 50 {
		return c.Status(400).JSON(fiber.Map{"error": "Username must be between 3 and 50 characters"})
	}

	// Validate role if provided
	if input.Role != "" {
		validRoles := map[string]bool{
			"user":        true,
			"instructor":  true,
			"admin":       true,
			"super_admin": true,
		}

		if !validRoles[strings.ToLower(input.Role)] {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid role. Valid roles are: user, instructor, admin, super_admin",
			})
		}
	}

	return c.Next()
}

// LoginValidator validates the input for the login route
func LoginValidator(c *fiber.Ctx) error {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	// Validate email and password
	if input.Email == "" || input.Password == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Email and password are required"})
	}
	return c.Next()
}

// ForgotPasswordValidator validates the input for the forgot-password route
func ForgotPasswordValidator(c *fiber.Ctx) error {
	var input struct {
		Email string `json:"email"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	// Validate email
	if input.Email == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Email is required"})
	}
	return c.Next()
}

// ResetPasswordValidator validates the input for the reset-password route
func ResetPasswordValidator(c *fiber.Ctx) error {
	var input struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	// Validate token and password
	if input.Token == "" || input.Password == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Reset token and new password are required"})
	}

	// Validate password length
	if len(input.Password) < 6 {
		return c.Status(400).JSON(fiber.Map{"error": "Password must be at least 6 characters long"})
	}

	return c.Next()
}
