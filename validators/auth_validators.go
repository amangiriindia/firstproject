package validators

import (
	"github.com/gofiber/fiber/v2"
)

// RegisterValidator validates the input for the registration route
func RegisterValidator(c *fiber.Ctx) error {
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
	return nil
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
	return nil
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
	return nil
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
	return nil
}
