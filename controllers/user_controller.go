package controllers

import (
	"firstproject/database"
	"firstproject/models"
	"github.com/gofiber/fiber/v2"
)

func GetProfile(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)

	// Custom response struct (excluding sensitive fields like PasswordHash)
	response := fiber.Map{
		"id":          user.ID,
		"username":    user.Username,
		"email":       user.Email,
		"first_name":  user.FirstName,
		"last_name":   user.LastName,
		"avatar_url":  user.AvatarURL,
		"bio":         user.Bio,
		"role":        user.Role,
		"is_verified": user.IsVerified,
		"profile":     user.Profile,
		"created_at":  user.CreatedAt,
		"updated_at":  user.UpdatedAt,
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  true,
		"message": "Profile fetched successfully",
		"data":    response,
	})
}

func UpdateProfile(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)

	type UpdateProfileInput struct {
		FirstName   string   `json:"first_name"`
		LastName    string   `json:"last_name"`
		AvatarURL   string   `json:"avatar_url"`
		Bio         string   `json:"bio"`
		Skills      []string `json:"skills"`    // JSON will unmarshal into []string
		Interests   []string `json:"interests"` // JSON will unmarshal into []string
		GithubURL   string   `json:"github_url"`
		LinkedinURL string   `json:"linkedin_url"`
		TwitterURL  string   `json:"twitter_url"`
		WebsiteURL  string   `json:"website_url"`
		Education   string   `json:"education"`
		Experience  string   `json:"experience"`
	}

	var input UpdateProfileInput

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  false,
			"message": "Invalid request body",
		})
	}

	// Update User fields
	user.FirstName = input.FirstName
	user.LastName = input.LastName
	user.AvatarURL = input.AvatarURL
	user.Bio = input.Bio

	// Update Profile fields
	profile := &user.Profile
	profile.Skills = input.Skills // Assigning []string to pq.StringArray works
	profile.Interests = input.Interests
	profile.GithubURL = input.GithubURL
	profile.LinkedinURL = input.LinkedinURL
	profile.TwitterURL = input.TwitterURL
	profile.WebsiteURL = input.WebsiteURL
	profile.Education = input.Education
	profile.Experience = input.Experience

	// Save both user and profile
	if err := database.DB.Save(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  false,
			"message": "Failed to update profile",
		})
	}

	if err := database.DB.Save(&profile).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  false,
			"message": "Failed to update profile details",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  true,
		"message": "Profile updated successfully",
		"data":    user,
	})
}
