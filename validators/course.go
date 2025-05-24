package validators

import (
	"firstproject/controllers"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"strings"
)

var validate1 = validator.New()

// CreateCourseValidator validates course creation input
func CreateCourseValidator(c *fiber.Ctx) error {
	var input controllers.CreateCourseInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  false,
			"message": "Invalid request body",
		})
	}

	// Validate struct
	if err := validate.Struct(&input); err != nil {
		var errors []string
		for _, err := range err.(validator.ValidationErrors) {
			errors = append(errors, getValidationError(err))
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  false,
			"message": "Validation failed",
			"errors":  errors,
		})
	}

	// Additional custom validations
	if input.Title != "" {
		input.Title = strings.TrimSpace(input.Title)
		if len(input.Title) < 3 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  false,
				"message": "Title must be at least 3 characters long",
			})
		}
	}

	if input.Currency != "" && len(input.Currency) != 3 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  false,
			"message": "Currency must be a valid 3-letter code (e.g., USD, EUR)",
		})
	}

	if input.Level != "" {
		validLevels := []string{"beginner", "intermediate", "advanced"}
		isValidLevel := false
		for _, level := range validLevels {
			if input.Level == level {
				isValidLevel = true
				break
			}
		}
		if !isValidLevel {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  false,
				"message": "Level must be one of: beginner, intermediate, advanced",
			})
		}
	}

	// Set defaults
	if input.Currency == "" {
		input.Currency = "USD"
	}
	if input.Level == "" {
		input.Level = "beginner"
	}
	if input.Language == "" {
		input.Language = "English"
	}

	c.Locals("input", input)
	return c.Next()
}

// UpdateCourseValidator validates course update input
func UpdateCourseValidator(c *fiber.Ctx) error {
	return CreateCourseValidator(c) // Same validation rules
}

// CreateContentValidator validates course content creation
func CreateContentValidator(c *fiber.Ctx) error {
	var input controllers.CreateContentInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  false,
			"message": "Invalid request body",
		})
	}

	if err := validate.Struct(&input); err != nil {
		var errors []string
		for _, err := range err.(validator.ValidationErrors) {
			errors = append(errors, getValidationError(err))
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  false,
			"message": "Validation failed",
			"errors":  errors,
		})
	}

	// Validate content type
	validTypes := []string{"mcq", "pdf", "text", "video", "image", "note", "assignment"}
	isValidType := false
	for _, t := range validTypes {
		if input.Type == t {
			isValidType = true
			break
		}
	}
	if !isValidType {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  false,
			"message": "Invalid content type. Must be one of: mcq, pdf, text, video, image, note, assignment",
		})
	}

	// Validate data based on content type
	if err := validateContentData(input.Type, input.Data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  false,
			"message": err.Error(),
		})
	}

	c.Locals("input", input)
	return c.Next()
}

// ReviewValidator validates review input
func ReviewValidator(c *fiber.Ctx) error {
	var input controllers.ReviewInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  false,
			"message": "Invalid request body",
		})
	}

	if err := validate.Struct(&input); err != nil {
		var errors []string
		for _, err := range err.(validator.ValidationErrors) {
			errors = append(errors, getValidationError(err))
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  false,
			"message": "Validation failed",
			"errors":  errors,
		})
	}

	if input.Rating < 1 || input.Rating > 5 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  false,
			"message": "Rating must be between 1 and 5",
		})
	}

	c.Locals("input", input)
	return c.Next()
}

// UpdateProgressValidator validates progress update input
func UpdateProgressValidator(c *fiber.Ctx) error {
	var input controllers.UpdateProgressInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  false,
			"message": "Invalid request body",
		})
	}

	if err := validate.Struct(&input); err != nil {
		var errors []string
		for _, err := range err.(validator.ValidationErrors) {
			errors = append(errors, getValidationError(err))
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  false,
			"message": "Validation failed",
			"errors":  errors,
		})
	}

	if input.ContentID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  false,
			"message": "Content ID is required",
		})
	}

	c.Locals("input", input)
	return c.Next()
}

// Helper function to get user-friendly validation error messages
func getValidationError(err validator.FieldError) string {
	field := err.Field()
	tag := err.Tag()

	switch tag {
	case "required":
		return field + " is required"
	case "min":
		return field + " must be at least " + err.Param() + " characters long"
	case "max":
		return field + " must be at most " + err.Param() + " characters long"
	case "url":
		return field + " must be a valid URL"
	case "email":
		return field + " must be a valid email address"
	case "len":
		return field + " must be exactly " + err.Param() + " characters long"
	case "oneof":
		return field + " must be one of: " + err.Param()
	default:
		return field + " is invalid"
	}
}

// validateContentData validates content data based on type
func validateContentData(contentType string, data interface{}) error {
	switch contentType {
	case "video":
		return validateVideoData(data)
	case "mcq":
		return validateMCQData(data)
	case "pdf":
		return validatePDFData(data)
	case "text":
		return validateTextData(data)
	case "image":
		return validateImageData(data)
	case "note":
		return validateNoteData(data)
	case "assignment":
		return validateAssignmentData(data)
	default:
		return nil
	}
}

func validateVideoData(data interface{}) error {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return fiber.NewError(fiber.StatusBadRequest, "Video data must be an object")
	}

	if url, exists := dataMap["url"]; !exists || url == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Video URL is required")
	}

	return nil
}

func validateMCQData(data interface{}) error {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return fiber.NewError(fiber.StatusBadRequest, "MCQ data must be an object")
	}

	if question, exists := dataMap["question"]; !exists || question == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Question is required for MCQ")
	}

	if options, exists := dataMap["options"]; !exists {
		return fiber.NewError(fiber.StatusBadRequest, "Options are required for MCQ")
	} else {
		optionsList, ok := options.([]interface{})
		if !ok || len(optionsList) < 2 {
			return fiber.NewError(fiber.StatusBadRequest, "MCQ must have at least 2 options")
		}
	}

	if correctAnswer, exists := dataMap["correct_answer"]; !exists {
		return fiber.NewError(fiber.StatusBadRequest, "Correct answer is required for MCQ")
	} else {
		if answer, ok := correctAnswer.(float64); !ok || answer < 0 {
			return fiber.NewError(fiber.StatusBadRequest, "Correct answer must be a valid option index")
		}
	}

	return nil
}

func validatePDFData(data interface{}) error {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return fiber.NewError(fiber.StatusBadRequest, "PDF data must be an object")
	}

	if url, exists := dataMap["url"]; !exists || url == "" {
		return fiber.NewError(fiber.StatusBadRequest, "PDF URL is required")
	}

	return nil
}

func validateTextData(data interface{}) error {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return fiber.NewError(fiber.StatusBadRequest, "Text data must be an object")
	}

	if content, exists := dataMap["content"]; !exists || content == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Text content is required")
	}

	return nil
}

func validateImageData(data interface{}) error {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return fiber.NewError(fiber.StatusBadRequest, "Image data must be an object")
	}

	if url, exists := dataMap["url"]; !exists || url == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Image URL is required")
	}

	return nil
}

func validateNoteData(data interface{}) error {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return fiber.NewError(fiber.StatusBadRequest, "Note data must be an object")
	}

	if content, exists := dataMap["content"]; !exists || content == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Note content is required")
	}

	return nil
}

func validateAssignmentData(data interface{}) error {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return fiber.NewError(fiber.StatusBadRequest, "Assignment data must be an object")
	}

	if title, exists := dataMap["title"]; !exists || title == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Assignment title is required")
	}

	if description, exists := dataMap["description"]; !exists || description == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Assignment description is required")
	}

	return nil
}
