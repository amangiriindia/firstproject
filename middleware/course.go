package middleware

import (
	"firstproject/database"
	"firstproject/models"
	"github.com/gofiber/fiber/v2"
)

func AuthorMiddleware(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	courseID := c.Params("id")

	var course models.Course
	if err := database.DB.First(&course, courseID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  false,
			"message": "Course not found",
		})
	}

	if course.AuthorID != user.ID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"status":  false,
			"message": "You are not the author of this course",
		})
	}

	c.Locals("course", &course)
	return c.Next()
}