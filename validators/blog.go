package validators

import (
	"firstproject/models"
	"github.com/gofiber/fiber/v2"
)

func CreateBlogValidator(c *fiber.Ctx) error {
	var input models.BlogInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	if input.Title == "" || input.Content == "" || input.CategoryID == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Title, content, and category are required"})
	}

	c.Locals("blogInput", input)
	return c.Next()
}
