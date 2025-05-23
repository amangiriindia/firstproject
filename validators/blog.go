package validators

import (
	"firstproject/models"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"strings"
)

var validate = validator.New()

// Blog validators
func CreateBlogValidator(c *fiber.Ctx) error {
	var input models.BlogInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	// Validate required fields
	if strings.TrimSpace(input.Title) == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Title is required"})
	}
	if strings.TrimSpace(input.Content) == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Content is required"})
	}
	if input.CategoryID == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Category ID is required"})
	}

	// Validate title length
	if len(input.Title) < 3 || len(input.Title) > 200 {
		return c.Status(400).JSON(fiber.Map{"error": "Title must be between 3 and 200 characters"})
	}

	// Validate content length
	if len(input.Content) < 10 {
		return c.Status(400).JSON(fiber.Map{"error": "Content must be at least 10 characters long"})
	}

	// Validate status
	if input.Status != "" && input.Status != "published" && input.Status != "draft" && input.Status != "archived" {
		return c.Status(400).JSON(fiber.Map{"error": "Status must be 'published', 'draft', or 'archived'"})
	}

	// Set default status if not provided
	if input.Status == "" {
		input.Status = "published"
	}

	// Clean up keywords
	if input.Keywords != "" {
		keywords := strings.Split(input.Keywords, ",")
		var cleanedKeywords []string
		for _, keyword := range keywords {
			keyword = strings.TrimSpace(keyword)
			if keyword != "" {
				cleanedKeywords = append(cleanedKeywords, keyword)
			}
		}
		input.Keywords = strings.Join(cleanedKeywords, ",")
	}

	c.Locals("blogInput", input)
	return c.Next()
}

// Category validators
func CreateCategoryValidator(c *fiber.Ctx) error {
	var input models.CategoryInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	// Validate required fields
	if strings.TrimSpace(input.Name) == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Category name is required"})
	}

	// Validate name length
	if len(input.Name) < 2 || len(input.Name) > 100 {
		return c.Status(400).JSON(fiber.Map{"error": "Category name must be between 2 and 100 characters"})
	}

	// Validate description length
	if len(input.Description) > 500 {
		return c.Status(400).JSON(fiber.Map{"error": "Description must be less than 500 characters"})
	}

	// Clean up input
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)

	c.Locals("categoryInput", input)
	return c.Next()
}

// Comment validators
func CreateCommentValidator(c *fiber.Ctx) error {
	var input models.CommentInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	// Validate required fields
	if strings.TrimSpace(input.Content) == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Comment content is required"})
	}
	if input.BlogID == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Blog ID is required"})
	}

	// Validate content length
	if len(input.Content) < 1 || len(input.Content) > 1000 {
		return c.Status(400).JSON(fiber.Map{"error": "Comment must be between 1 and 1000 characters"})
	}

	// Clean up content
	input.Content = strings.TrimSpace(input.Content)

	c.Locals("commentInput", input)
	return c.Next()
}

// Search parameter validator
func SearchParamsValidator(c *fiber.Ctx) error {
	var params models.BlogSearchParams

	// Parse query parameters
	params.Query = strings.TrimSpace(c.Query("q"))
	params.CategoryID = uint(c.QueryInt("category_id", 0))
	params.AuthorID = uint(c.QueryInt("author_id", 0))
	params.Status = c.Query("status", "published")
	params.SortBy = c.Query("sort_by", "created_at")
	params.Order = c.Query("order", "desc")
	params.Page = c.QueryInt("page", 1)
	params.Limit = c.QueryInt("limit", 10)

	// Validate sort_by
	validSortFields := []string{"created_at", "updated_at", "view_count", "title"}
	isValidSort := false
	for _, field := range validSortFields {
		if params.SortBy == field {
			isValidSort = true
			break
		}
	}
	if !isValidSort {
		params.SortBy = "created_at"
	}

	// Validate order
	if params.Order != "asc" && params.Order != "desc" {
		params.Order = "desc"
	}

	// Validate status
	if params.Status != "" && params.Status != "published" && params.Status != "draft" && params.Status != "archived" {
		params.Status = "published"
	}

	// Validate pagination
	if params.Page < 1 {
		params.Page = 1
	}
	if params.Limit < 1 || params.Limit > 100 {
		params.Limit = 10
	}

	c.Locals("searchParams", params)
	return c.Next()
}
