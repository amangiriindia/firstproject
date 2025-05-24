package middleware

import (
	"firstproject/database"
	"firstproject/models"
	"strconv"

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

// EnrollmentMiddleware checks if user is enrolled in the course
func EnrollmentMiddleware(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	courseIDStr := c.Params("courseId")

	// Convert string courseId to uint
	courseID, err := strconv.ParseUint(courseIDStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  false,
			"message": "Invalid course ID format",
		})
	}

	var enrollment models.Enrollment
	if err := database.DB.Where("user_id = ? AND course_id = ?", user.ID, uint(courseID)).
		First(&enrollment).Error; err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"status":  false,
			"message": "You are not enrolled in this course",
		})
	}

	c.Locals("enrollment", &enrollment)
	return c.Next()
}

// ContentMiddleware validates content ownership
func ContentMiddleware(c *fiber.Ctx) error {
	contentID := c.Params("contentId")
	course := c.Locals("course").(*models.Course)

	var content models.CourseContent
	if err := database.DB.First(&content, contentID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  false,
			"message": "Content not found",
		})
	}

	if content.CourseID != course.ID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"status":  false,
			"message": "Content does not belong to this course",
		})
	}

	c.Locals("content", &content)
	return c.Next()
}
