package controllers

import (
	"encoding/json"
	"firstproject/database"
	"firstproject/models"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Input structs for validation
type CreateCourseInput struct {
	Title         string   `json:"title" validate:"required,min=3,max=255"`
	Description   string   `json:"description" validate:"max=5000"`
	FeaturedImage string   `json:"featured_image" validate:"url"`
	Price         float64  `json:"price" validate:"min=0"`
	Currency      string   `json:"currency" validate:"len=3"`
	Level         string   `json:"level" validate:"oneof=beginner intermediate advanced"`
	Duration      int      `json:"duration" validate:"min=0"`
	Language      string   `json:"language" validate:"required,max=50"`
	Category      string   `json:"category" validate:"required,max=100"`
	Tags          []string `json:"tags"`
}

type CreateContentInput struct {
	Title     string      `json:"title" validate:"required,min=1,max=255"`
	Type      string      `json:"type" validate:"required,oneof=mcq pdf text video image note assignment"`
	Data      interface{} `json:"data" validate:"required"`
	Duration  int         `json:"duration" validate:"min=0"`
	Order     int         `json:"order" validate:"required,min=1"`
	IsPreview bool        `json:"is_preview"`
}

type UpdateProgressInput struct {
	ContentID    uint `json:"content_id" validate:"required"`
	IsCompleted  bool `json:"is_completed"`
	TimeSpent    int  `json:"time_spent" validate:"min=0"`
	LastPosition int  `json:"last_position" validate:"min=0"`
}

type ReviewInput struct {
	Rating  int    `json:"rating" validate:"required,min=1,max=5"`
	Comment string `json:"comment" validate:"max=1000"`
}

// GetCourses lists all courses with filtering and pagination
func GetCourses(c *fiber.Ctx) error {
	var courses []models.Course
	query := database.DB.Preload("Author").Where("is_published = ?", true)

	// Filtering
	if category := c.Query("category"); category != "" {
		query = query.Where("category = ?", category)
	}
	if level := c.Query("level"); level != "" {
		query = query.Where("level = ?", level)
	}
	if search := c.Query("search"); search != "" {
		query = query.Where("title ILIKE ? OR description ILIKE ?", "%"+search+"%", "%"+search+"%")
	}
	if minPrice := c.Query("min_price"); minPrice != "" {
		if price, err := strconv.ParseFloat(minPrice, 64); err == nil {
			query = query.Where("price >= ?", price)
		}
	}
	if maxPrice := c.Query("max_price"); maxPrice != "" {
		if price, err := strconv.ParseFloat(maxPrice, 64); err == nil {
			query = query.Where("price <= ?", price)
		}
	}

	// Sorting
	sortBy := c.Query("sort", "created_at")
	order := c.Query("order", "desc")
	if order != "asc" && order != "desc" {
		order = "desc"
	}
	query = query.Order(fmt.Sprintf("%s %s", sortBy, order))

	// Pagination
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	var total int64
	query.Model(&models.Course{}).Count(&total)

	if err := query.Offset(offset).Limit(limit).Find(&courses).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  false,
			"message": "Failed to fetch courses",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  true,
		"message": "Courses fetched successfully",
		"data": fiber.Map{
			"courses": courses,
			"pagination": fiber.Map{
				"page":  page,
				"limit": limit,
				"total": total,
				"pages": (total + int64(limit) - 1) / int64(limit),
			},
		},
	})
}

// GetCourse fetches a single course by ID
func GetCourse(c *fiber.Ctx) error {
	courseID := c.Params("id")
	var course models.Course

	query := database.DB.Preload("Author").Preload("Contents", func(db *gorm.DB) *gorm.DB {
		return db.Order("\"order\" ASC")
	})

	if err := query.First(&course, courseID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  false,
			"message": "Course not found",
		})
	}

	// Check if user is enrolled (if authenticated)
	var isEnrolled bool
	if user := c.Locals("user"); user != nil {
		userModel := user.(*models.User)
		var enrollment models.Enrollment
		if err := database.DB.Where("user_id = ? AND course_id = ?", userModel.ID, course.ID).First(&enrollment).Error; err == nil {
			isEnrolled = true
		}
	}

	// Filter contents based on enrollment and preview status
	if !isEnrolled {
		var previewContents []models.CourseContent
		for _, content := range course.Contents {
			if content.IsPreview {
				previewContents = append(previewContents, content)
			}
		}
		course.Contents = previewContents
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":   true,
		"message":  "Course fetched successfully",
		"data":     course,
		"enrolled": isEnrolled,
	})
}

// CreateCourse creates a new course
func CreateCourse(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	input := c.Locals("input").(CreateCourseInput)

	// Convert data to JSON string if needed
	// dataJSON, err := json.Marshal(input.Tags)
	// if err != nil {
	// 	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
	// 		"status":  false,
	// 		"message": "Failed to process tags",
	// 	})
	// }

	course := models.Course{
		Title:         input.Title,
		Description:   input.Description,
		FeaturedImage: input.FeaturedImage,
		Price:         input.Price,
		Currency:      input.Currency,
		Level:         input.Level,
		Duration:      input.Duration,
		Language:      input.Language,
		Category:      input.Category,
		Tags:          input.Tags,
		AuthorID:      user.ID,
		IsPublished:   false,
	}

	if err := database.DB.Create(&course).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  false,
			"message": "Failed to create course",
		})
	}

	// Load author information
	database.DB.Preload("Author").First(&course, course.ID)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status":  true,
		"message": "Course created successfully",
		"data":    course,
	})
}

// UpdateCourse updates an existing course
func UpdateCourse(c *fiber.Ctx) error {
	course := c.Locals("course").(*models.Course)
	input := c.Locals("input").(CreateCourseInput)

	course.Title = input.Title
	course.Description = input.Description
	course.FeaturedImage = input.FeaturedImage
	course.Price = input.Price
	course.Currency = input.Currency
	course.Level = input.Level
	course.Duration = input.Duration
	course.Language = input.Language
	course.Category = input.Category
	course.Tags = input.Tags

	if err := database.DB.Save(&course).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  false,
			"message": "Failed to update course",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  true,
		"message": "Course updated successfully",
		"data":    course,
	})
}

// DeleteCourse deletes a course
func DeleteCourse(c *fiber.Ctx) error {
	course := c.Locals("course").(*models.Course)

	if err := database.DB.Delete(&course).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  false,
			"message": "Failed to delete course",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  true,
		"message": "Course deleted successfully",
	})
}

// PublishCourse publishes a course
func PublishCourse(c *fiber.Ctx) error {
	course := c.Locals("course").(*models.Course)

	course.IsPublished = true
	if err := database.DB.Save(&course).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  false,
			"message": "Failed to publish course",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  true,
		"message": "Course published successfully",
		"data":    course,
	})
}

// EnrollCourse enrolls a user in a course
func EnrollCourse(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	courseID := c.Params("id")

	var existingEnrollment models.Enrollment
	if err := database.DB.Where("user_id = ? AND course_id = ?", user.ID, courseID).First(&existingEnrollment).Error; err == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  false,
			"message": "You are already enrolled in this course",
		})
	}

	var course models.Course
	if err := database.DB.First(&course, courseID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  false,
			"message": "Course not found",
		})
	}

	enrollment := models.Enrollment{
		UserID:        user.ID,
		CourseID:      course.ID,
		EnrolledAt:    time.Now(),
		Progress:      0,
		PaymentStatus: "pending",
	}

	// For free courses, mark payment as completed
	if course.Price == 0 {
		enrollment.PaymentStatus = "completed"
	}

	if err := database.DB.Create(&enrollment).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  false,
			"message": "Failed to enroll in course",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status":  true,
		"message": "Enrolled in course successfully",
		"data":    enrollment,
	})
}

// GetCourseContent fetches course content for enrolled users
func GetCourseContent(c *fiber.Ctx) error {
	courseID := c.Params("id")
	enrollment := c.Locals("enrollment").(*models.Enrollment)

	var contents []models.CourseContent
	if err := database.DB.Where("course_id = ?", courseID).Order("\"order\" ASC").Find(&contents).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  false,
			"message": "Failed to fetch course content",
		})
	}

	// Get content progress
	var progress []models.ContentProgress
	database.DB.Where("enrollment_id = ?", enrollment.ID).Find(&progress)

	// Map progress to contents
	progressMap := make(map[uint]models.ContentProgress)
	for _, p := range progress {
		progressMap[p.ContentID] = p
	}

	// Add progress info to contents
	for i := range contents {
		if prog, exists := progressMap[contents[i].ID]; exists {
			contents[i].IsCompleted = prog.IsCompleted
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  true,
		"message": "Course content fetched successfully",
		"data":    contents,
	})
}

// UpdateProgress updates user's progress for a specific content
func UpdateProgress(c *fiber.Ctx) error {
	enrollment := c.Locals("enrollment").(*models.Enrollment)
	var input UpdateProgressInput

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  false,
			"message": "Invalid request body",
		})
	}

	// Find or create content progress
	var contentProgress models.ContentProgress
	if err := database.DB.Where("enrollment_id = ? AND content_id = ?", enrollment.ID, input.ContentID).First(&contentProgress).Error; err != nil {
		contentProgress = models.ContentProgress{
			EnrollmentID: enrollment.ID,
			ContentID:    input.ContentID,
		}
	}

	contentProgress.IsCompleted = input.IsCompleted
	contentProgress.TimeSpent += input.TimeSpent
	contentProgress.LastPosition = input.LastPosition

	if input.IsCompleted && contentProgress.CompletedAt == nil {
		now := time.Now()
		contentProgress.CompletedAt = &now
	}

	if err := database.DB.Save(&contentProgress).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  false,
			"message": "Failed to update progress",
		})
	}

	// Update overall enrollment progress
	updateEnrollmentProgress(enrollment.ID)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  true,
		"message": "Progress updated successfully",
		"data":    contentProgress,
	})
}

// CompleteCourse marks a course as completed and issues a certificate
func CompleteCourse(c *fiber.Ctx) error {
	enrollment := c.Locals("enrollment").(*models.Enrollment)

	if enrollment.CompletedAt != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  false,
			"message": "Course already completed",
		})
	}

	now := time.Now()
	enrollment.CompletedAt = &now
	enrollment.Progress = 100

	if err := database.DB.Save(&enrollment).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  false,
			"message": "Failed to mark course as completed",
		})
	}

	certificate := models.Certificate{
		UserID:        enrollment.UserID,
		CourseID:      enrollment.CourseID,
		CertificateID: generateCertificateID(),
		IssuedAt:      now,
	}

	if err := database.DB.Create(&certificate).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  false,
			"message": "Failed to issue certificate",
		})
	}

	// Load related data
	database.DB.Preload("User").Preload("Course").First(&certificate, certificate.ID)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  true,
		"message": "Course completed and certificate issued",
		"data":    certificate,
	})
}

// GetCertificates lists a user's certificates
func GetCertificates(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	var certificates []models.Certificate

	if err := database.DB.
		Preload("User").                  // Preload user data
		Preload("User.Profile").          // Preload user profile
		Preload("Course").                // Preload course data
		Preload("Course.Author").         // Preload course author
		Preload("Course.Author.Profile"). // Preload course author profile
		Where("user_id = ?", user.ID).
		Find(&certificates).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  false,
			"message": "Failed to fetch certificates",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  true,
		"message": "Certificates fetched successfully",
		"data":    certificates,
	})
}

// AddCourseContent adds content to a course
func AddCourseContent(c *fiber.Ctx) error {
	course := c.Locals("course").(*models.Course)
	var input CreateContentInput

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  false,
			"message": "Invalid request body",
		})
	}

	// Convert data to JSON string
	dataJSON, err := json.Marshal(input.Data)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  false,
			"message": "Failed to process content data",
		})
	}

	content := models.CourseContent{
		CourseID:  course.ID,
		Title:     input.Title,
		Type:      input.Type,
		Data:      string(dataJSON),
		Duration:  input.Duration,
		Order:     input.Order,
		IsPreview: input.IsPreview,
	}

	if err := database.DB.Create(&content).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  false,
			"message": "Failed to add content",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status":  true,
		"message": "Content added successfully",
		"data":    content,
	})
}

// AddReview adds a review for a course
func AddReview(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	courseID := c.Params("id")
	var input ReviewInput

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  false,
			"message": "Invalid request body",
		})
	}

	// Check if user is enrolled and has completed the course
	var enrollment models.Enrollment
	if err := database.DB.Where("user_id = ? AND course_id = ? AND completed_at IS NOT NULL", user.ID, courseID).First(&enrollment).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  false,
			"message": "You must complete the course before reviewing",
		})
	}

	review := models.Review{
		UserID:   user.ID,
		CourseID: enrollment.CourseID,
		Rating:   input.Rating,
		Comment:  input.Comment,
	}

	if err := database.DB.Create(&review).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  false,
			"message": "Failed to add review",
		})
	}

	// Load user data
	database.DB.Preload("User").First(&review, review.ID)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status":  true,
		"message": "Review added successfully",
		"data":    review,
	})
}

// Helper functions
func updateEnrollmentProgress(enrollmentID uint) {
	var enrollment models.Enrollment
	database.DB.First(&enrollment, enrollmentID)

	var totalContents int64
	var completedContents int64

	database.DB.Model(&models.CourseContent{}).Where("course_id = ?", enrollment.CourseID).Count(&totalContents)
	database.DB.Model(&models.ContentProgress{}).Where("enrollment_id = ? AND is_completed = ?", enrollmentID, true).Count(&completedContents)

	if totalContents > 0 {
		progress := int((completedContents * 100) / totalContents)
		enrollment.Progress = progress
		database.DB.Save(&enrollment)
	}
}

func generateCertificateID() string {
	return fmt.Sprintf("CERT-%s", uuid.New().String()[:8])
}

// Additional controller functions for enhanced course management

// UpdateCourseContent updates existing course content
// func UpdateCourseContent(c *fiber.Ctx) error {
// 	course := c.Locals("course").(*models.Course)
// 	contentID := c.Params("contentId")
// 	input := c.Locals("input").(CreateContentInput)

// 	var content models.CourseContent
// 	if err := database.DB.Where("id = ? AND course_id = ?", contentID, course.ID).First(&content).Error; err != nil {
// 		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
// 			"status":  false,
// 			"message": "Content not found",
// 		})
// 	}

// 	dataJSON, err := json.Marshal(input.Data)
// 	if err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"status":  false,
// 			"message": "Failed to process content data",
// 		})
// 	}

// 	content.Title = input.Title
// 	content.Type = input.Type
// 	content.Data = string(dataJSON)
// 	content.Duration = input.Duration
// 	content.Order = input.Order
// 	content.IsPreview = input.IsPreview

// 	if err := database.DB.Save(&content).Error; err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"status":  false,
// 			"message": "Failed to update content",
// 		})
// 	}

// 	return c.Status(fiber.StatusOK).JSON(fiber.Map{
// 		"status":  true,
// 		"message": "Content updated successfully",
// 		"data":    content,
// 	})
// }

// // DeleteCourseContent deletes course content
// func DeleteCourseContent(c *fiber.Ctx) error {
// 	course := c.Locals("course").(*models.Course)
// 	contentID := c.Params("contentId")

// 	var content models.CourseContent
// 	if err := database.DB.Where("id = ? AND course_id = ?", contentID, course.ID).First(&content).Error; err != nil {
// 		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
// 			"status":  false,
// 			"message": "Content not found",
// 		})
// 	}

// 	if err := database.DB.Delete(&content).Error; err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"status":  false,
// 			"message": "Failed to delete content",
// 		})
// 	}

// 	return c.Status(fiber.StatusOK).JSON(fiber.Map{
// 		"status":  true,
// 		"message": "Content deleted successfully",
// 	})
// }

// GetCourseContentForAuthor gets all content for course author
func GetCourseContentForAuthor(c *fiber.Ctx) error {
	course := c.Locals("course").(*models.Course)

	var contents []models.CourseContent
	if err := database.DB.Where("course_id = ?", course.ID).Order("\"order\" ASC").Find(&contents).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  false,
			"message": "Failed to fetch course content",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  true,
		"message": "Course content fetched successfully",
		"data":    contents,
	})
}

// UnenrollCourse removes user from course
func UnenrollCourse(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)
	courseID := c.Params("id")

	var enrollment models.Enrollment
	if err := database.DB.Where("user_id = ? AND course_id = ?", user.ID, courseID).First(&enrollment).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  false,
			"message": "Enrollment not found",
		})
	}

	// Delete associated progress
	database.DB.Where("enrollment_id = ?", enrollment.ID).Delete(&models.ContentProgress{})

	if err := database.DB.Delete(&enrollment).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  false,
			"message": "Failed to unenroll from course",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  true,
		"message": "Successfully unenrolled from course",
	})
}

// GetSingleContent gets specific content for enrolled student
func GetSingleContent(c *fiber.Ctx) error {
	courseID := c.Params("id")
	contentID := c.Params("contentId")
	enrollment := c.Locals("enrollment").(*models.Enrollment)

	var content models.CourseContent
	if err := database.DB.Where("id = ? AND course_id = ?", contentID, courseID).First(&content).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  false,
			"message": "Content not found",
		})
	}

	// Get progress for this content
	var progress models.ContentProgress
	database.DB.Where("enrollment_id = ? AND content_id = ?", enrollment.ID, content.ID).First(&progress)

	content.IsCompleted = progress.IsCompleted

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":   true,
		"message":  "Content fetched successfully",
		"data":     content,
		"progress": progress,
	})
}

// GetNextContent gets next unfinished content
// func GetNextContent(c *fiber.Ctx) error {
// 	courseID := c.Params("id")
// 	enrollment := c.Locals("enrollment").(*models.Enrollment)

// 	// Get all contents for the course
// 	var contents []models.CourseContent
// 	database.DB.Where("course_id = ?", courseID).Order("\"order\" ASC").Find(&contents)

// 	// Get completed content IDs
// 	var completedProgress []models.ContentProgress
// 	database.DB.Where("enrollment_id = ? AND is_completed = ?", enrollment.ID, true).Find(&completedProgress)

// 	completedMap := make(map[uint]bool)
// 	for _, p := range completedProgress {
// 		completedMap[p.ContentID] = true
// 	}

// 	// Find first uncompleted content
// 	for _, content := range contents {
// 		if !completedMap[content.ID] {
// 			return c.Status(fiber.StatusOK).JSON(fiber.Map{
// 				"status":  true,
// 				"message": "Next content found",
// 				"data":    content,
// 			})
// 		}
// 	}

// 	return c.Status(fiber.StatusOK).JSON(fiber.Map{
// 		"status":  true,
// 		"message": "All content completed",
// 		"data":    nil,
// 	})
// }

// // GetCourseProgress gets detailed progress for a course
// func GetCourseProgress(c *fiber.Ctx) error {
// 	courseID := c.Params("id")
// 	enrollment := c.Locals("enrollment").(*models.Enrollment)

// 	var contents []models.CourseContent
// 	database.DB.Where("course_id = ?", courseID).Order("\"order\" ASC").Find(&contents)

// 	var progress []models.ContentProgress
// 	database.DB.Where("enrollment_id = ?", enrollment.ID).Find(&progress)

// 	progressMap := make(map[uint]models.ContentProgress)
// 	for _, p := range progress {
// 		progressMap[p.ContentID] = p
// 	}

// 	var detailedProgress []map[string]interface{}
// 	totalTimeSpent := 0
// 	completedCount := 0

// 	for _, content := range contents {
// 		prog, exists := progressMap[content.ID]
// 		if exists {
// 			totalTimeSpent += prog.TimeSpent
// 			if prog.IsCompleted {
// 				completedCount++
// 			}
// 		}

// 		detailedProgress = append(detailedProgress, map[string]interface{}{
// 			"content_id":   content.ID,
// 			"title":        content.Title,
// 			"type":         content.Type,
// 			"order":        content.Order,
// 			"duration":     content.Duration,
// 			"is_completed": exists && prog.IsCompleted,
// 			"time_spent": func() int {
// 				if exists {
// 					return prog.TimeSpent
// 				} else {
// 					return 0
// 				}
// 			}(),
// 			"last_position": func() int {
// 				if exists {
// 					return prog.LastPosition
// 				} else {
// 					return 0
// 				}
// 			}(),
// 			"completed_at": func() *time.Time {
// 				if exists {
// 					return prog.CompletedAt
// 				} else {
// 					return nil
// 				}
// 			}(),
// 		})
// 	}

// 	progressPercentage := 0
// 	if len(contents) > 0 {
// 		progressPercentage = (completedCount * 100) / len(contents)
// 	}

// 	return c.Status(fiber.StatusOK).JSON(fiber.Map{
// 		"status":  true,
// 		"message": "Progress fetched successfully",
// 		"data": map[string]interface{}{
// 			"course_id":           courseID,
// 			"progress_percentage": progressPercentage,
// 			"completed_contents":  completedCount,
// 			"total_contents":      len(contents),
// 			"total_time_spent":    totalTimeSpent,
// 			"enrolled_at":         enrollment.EnrolledAt,
// 			"completed_at":        enrollment.CompletedAt,
// 			"detailed_progress":   detailedProgress,
// 		},
// 	})
// }

// // GetEnrolledCourses gets all courses user is enrolled in
// func GetEnrolledCourses(c *fiber.Ctx) error {
// 	user := c.Locals("user").(*models.User)

// 	var enrollments []models.Enrollment
// 	if err := database.DB.Preload("Course").Preload("Course.Author").Where("user_id = ?", user.ID).Find(&enrollments).Error; err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"status":  false,
// 			"message": "Failed to fetch enrolled courses",
// 		})
// 	}

// 	return c.Status(fiber.StatusOK).JSON(fiber.Map{
// 		"status":  true,
// 		"message": "Enrolled courses fetched successfully",
// 		"data":    enrollments,
// 	})
// }

// GetAllCoursesProgress gets progress for all enrolled courses
func GetAllCoursesProgress(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)

	var enrollments []models.Enrollment
	database.DB.Preload("Course").Where("user_id = ?", user.ID).Find(&enrollments)

	var coursesProgress []map[string]interface{}

	for _, enrollment := range enrollments {
		var totalContents int64
		var completedContents int64

		database.DB.Model(&models.CourseContent{}).Where("course_id = ?", enrollment.CourseID).Count(&totalContents)
		database.DB.Model(&models.ContentProgress{}).Where("enrollment_id = ? AND is_completed = ?", enrollment.ID, true).Count(&completedContents)

		progressPercentage := 0
		if totalContents > 0 {
			progressPercentage = int((completedContents * 100) / totalContents)
		}

		coursesProgress = append(coursesProgress, map[string]interface{}{
			"course_id":           enrollment.CourseID,
			"course_title":        enrollment.Course.Title,
			"progress_percentage": progressPercentage,
			"completed_contents":  completedContents,
			"total_contents":      totalContents,
			"enrolled_at":         enrollment.EnrolledAt,
			"completed_at":        enrollment.CompletedAt,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  true,
		"message": "Courses progress fetched successfully",
		"data":    coursesProgress,
	})
}

// GetRecentActivity gets user's recent learning activity
func GetRecentActivity(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)

	var recentProgress []models.ContentProgress
	database.DB.Preload("Content").Preload("Content.Course").Where("enrollment_id IN (SELECT id FROM enrollments WHERE user_id = ?)", user.ID).Order("updated_at DESC").Limit(10).Find(&recentProgress)

	var activities []map[string]interface{}
	for _, progress := range recentProgress {
		activities = append(activities, map[string]interface{}{
			"type": "content_progress",
			// "course_title": progress.Content.Course.Title,
			// "content_title": progress.Content.Title,
			"is_completed": progress.IsCompleted,
			"time_spent":   progress.TimeSpent,
			"updated_at":   progress.UpdatedAt,
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  true,
		"message": "Recent activity fetched successfully",
		"data":    activities,
	})
}

// GetAchievements gets user achievements and stats
func GetAchievements(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)

	var totalEnrollments int64
	var completedCourses int64
	var totalCertificates int64
	var totalTimeSpent int64

	database.DB.Model(&models.Enrollment{}).Where("user_id = ?", user.ID).Count(&totalEnrollments)
	database.DB.Model(&models.Enrollment{}).Where("user_id = ? AND completed_at IS NOT NULL", user.ID).Count(&completedCourses)
	database.DB.Model(&models.Certificate{}).Where("user_id = ?", user.ID).Count(&totalCertificates)

	// Calculate total time spent
	database.DB.Model(&models.ContentProgress{}).Select("COALESCE(SUM(time_spent), 0)").Where("enrollment_id IN (SELECT id FROM enrollments WHERE user_id = ?)", user.ID).Scan(&totalTimeSpent)

	achievements := map[string]interface{}{
		"total_enrollments":  totalEnrollments,
		"completed_courses":  completedCourses,
		"total_certificates": totalCertificates,
		"total_time_spent":   totalTimeSpent,
		"badges":             []map[string]interface{}{},
	}

	// Add badges based on achievements
	if completedCourses >= 1 {
		achievements["badges"] = append(achievements["badges"].([]map[string]interface{}), map[string]interface{}{
			"name":        "First Course Completed",
			"description": "Completed your first course",
			"icon":        "🎓",
		})
	}
	if completedCourses >= 5 {
		achievements["badges"] = append(achievements["badges"].([]map[string]interface{}), map[string]interface{}{
			"name":        "Learning Enthusiast",
			"description": "Completed 5 courses",
			"icon":        "📚",
		})
	}
	if totalTimeSpent >= 3600 { // 1 hour in seconds
		achievements["badges"] = append(achievements["badges"].([]map[string]interface{}), map[string]interface{}{
			"name":        "Dedicated Learner",
			"description": "Spent over 1 hour learning",
			"icon":        "⏰",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  true,
		"message": "Achievements fetched successfully",
		"data":    achievements,
	})
}

// GetCourseReviews gets all reviews for a course
// func GetCourseReviews(c *fiber.Ctx) error {
// 	courseID := c.Params("id")

// 	var reviews []models.Review
// 	if err := database.DB.Preload("User").Where("course_id = ?", courseID).Order("created_at DESC").Find(&reviews).Error; err != nil {
// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
// 			"status":  false,
// 			"message": "Failed to fetch reviews",
// 		})
// 	}

// 	// Calculate average rating
// 	var avgRating float64
// 	if len(reviews) > 0 {
// 		totalRating := 0
// 		for _, review := range reviews {
// 			totalRating += review.Rating
// 		}
// 		avgRating = float64(totalRating) / float64(len(reviews))
// 	}

// 	return c.Status(fiber.StatusOK).JSON(fiber.Map{
// 		"status":  true,
// 		"message": "Reviews fetched successfully",
// 		"data": map[string]interface{}{
// 			"reviews":        reviews,
// 			"total_reviews":  len(reviews),
// 			"average_rating": avgRating,
// 		},
// 	})
// }

// UpdateReview updates user's review
func UpdateReview(c *fiber.Ctx) error {
	review := c.Locals("review").(*models.Review)
	input := c.Locals("input").(ReviewInput)

	review.Rating = input.Rating
	review.Comment = input.Comment

	if err := database.DB.Save(&review).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  false,
			"message": "Failed to update review",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  true,
		"message": "Review updated successfully",
		"data":    review,
	})
}

// DeleteReview deletes user's review
func DeleteReview(c *fiber.Ctx) error {
	review := c.Locals("review").(*models.Review)

	if err := database.DB.Delete(&review).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  false,
			"message": "Failed to delete review",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  true,
		"message": "Review deleted successfully",
	})
}

// controllers/course.go
// GetEnrolledCourses lists all enrolled courses for a user
func GetEnrolledCourses(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)

	var enrollments []models.Enrollment
	if err := database.DB.
		Preload("User").                  // Preload user data
		Preload("User.Profile").          // Preload user profile
		Preload("Course").                // Preload course data
		Preload("Course.Author").         // Preload course author
		Preload("Course.Author.Profile"). // Preload course author profile
		Where("user_id = ?", user.ID).
		Find(&enrollments).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  false,
			"message": "Failed to fetch enrolled courses",
		})
	}

	return c.JSON(fiber.Map{
		"status": true,
		"data":   enrollments,
	})
}

// GetEnrolledCourse gets details of a specific enrolled course
func GetEnrolledCourse(c *fiber.Ctx) error {
	enrollment := c.Locals("enrollment").(*models.Enrollment)

	// The enrollment from middleware doesn't have preloaded associations
	// So we need to fetch it again with proper preloading
	var fullEnrollment models.Enrollment
	if err := database.DB.
		Preload("User").                  // Preload user data
		Preload("User.Profile").          // Preload user profile
		Preload("Course").                // Preload course data
		Preload("Course.Author").         // Preload course author
		Preload("Course.Author.Profile"). // Preload course author profile
		Where("id = ?", enrollment.ID).
		First(&fullEnrollment).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  false,
			"message": "Failed to fetch enrollment details",
		})
	}

	return c.JSON(fiber.Map{
		"status": true,
		"data":   fullEnrollment,
	})
}

// // GetEnrolledCourse gets details of a specific enrolled course
// func GetEnrolledCourse(c *fiber.Ctx) error {
// 	enrollment := c.Locals("enrollment").(*models.Enrollment)
// 	return c.JSON(fiber.Map{
// 		"status": true,
// 		"data":   enrollment,
// 	})
// }

// GetCourseProgress gets detailed progress for a course
func GetCourseProgress(c *fiber.Ctx) error {
	enrollment := c.Locals("enrollment").(*models.Enrollment)

	var progress []models.ContentProgress
	if err := database.DB.Where("enrollment_id = ?", enrollment.ID).
		Find(&progress).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  false,
			"message": "Failed to fetch progress",
		})
	}

	return c.JSON(fiber.Map{
		"status": true,
		"data": fiber.Map{
			"overall":   enrollment.Progress,
			"details":   progress,
			"completed": enrollment.CompletedAt != nil,
		},
	})
}

// GetContentWithProgress gets specific content with user's progress
func GetContentWithProgress(c *fiber.Ctx) error {
	enrollment := c.Locals("enrollment").(*models.Enrollment)
	contentID := c.Params("contentId")

	var content models.CourseContent
	if err := database.DB.First(&content, contentID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  false,
			"message": "Content not found",
		})
	}

	var progress models.ContentProgress
	database.DB.Where("enrollment_id = ? AND content_id = ?",
		enrollment.ID, contentID).First(&progress)

	return c.JSON(fiber.Map{
		"status": true,
		"data": fiber.Map{
			"content":  content,
			"progress": progress,
		},
	})
}

// GetNextContent gets the next uncompleted content in order
func GetNextContent(c *fiber.Ctx) error {
	enrollment := c.Locals("enrollment").(*models.Enrollment)

	var nextContent models.CourseContent
	query := database.DB.Joins("LEFT JOIN content_progresses ON content_progresses.content_id = course_contents.id AND content_progresses.enrollment_id = ?", enrollment.ID).
		Where("course_contents.course_id = ? AND (content_progresses.is_completed = false OR content_progresses.id IS NULL)", enrollment.CourseID).
		Order("course_contents.order ASC"). // Fixed: removed quotes around "order"
		First(&nextContent)

	if query.Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  false,
			"message": "No more content to complete",
		})
	}

	return c.JSON(fiber.Map{
		"status": true,
		"data":   nextContent,
	})
}

// GetResumePosition gets the last accessed content for resuming
func GetResumePosition(c *fiber.Ctx) error {
	enrollment := c.Locals("enrollment").(*models.Enrollment)

	var lastProgress models.ContentProgress
	if err := database.DB.Where("enrollment_id = ?", enrollment.ID).
		Order("updated_at DESC").
		First(&lastProgress).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  false,
			"message": "No progress found",
		})
	}

	var content models.CourseContent
	database.DB.First(&content, lastProgress.ContentID)

	return c.JSON(fiber.Map{
		"status": true,
		"data": fiber.Map{
			"content":       content,
			"last_position": lastProgress.LastPosition,
		},
	})
}

// UpdateCourseContent updates existing course content (for instructors)
func UpdateCourseContent(c *fiber.Ctx) error {
	contentID := c.Params("contentId")
	var content models.CourseContent
	if err := database.DB.First(&content, contentID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  false,
			"message": "Content not found",
		})
	}

	var input CreateContentInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  false,
			"message": "Invalid request body",
		})
	}

	// Update content fields
	content.Title = input.Title
	content.Type = input.Type
	content.Duration = input.Duration
	content.Order = input.Order
	content.IsPreview = input.IsPreview

	// Update data
	dataJSON, err := json.Marshal(input.Data)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  false,
			"message": "Failed to process content data",
		})
	}
	content.Data = string(dataJSON)

	if err := database.DB.Save(&content).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  false,
			"message": "Failed to update content",
		})
	}

	return c.JSON(fiber.Map{
		"status": true,
		"data":   content,
	})
}

// DeleteCourseContent deletes course content (for instructors)
func DeleteCourseContent(c *fiber.Ctx) error {
	contentID := c.Params("contentId")
	var content models.CourseContent
	if err := database.DB.First(&content, contentID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  false,
			"message": "Content not found",
		})
	}

	if err := database.DB.Delete(&content).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  false,
			"message": "Failed to delete content",
		})
	}

	return c.JSON(fiber.Map{
		"status":  true,
		"message": "Content deleted successfully",
	})
}

// GetAllCourseContent lists all content for course management
func GetAllCourseContent(c *fiber.Ctx) error {
	course := c.Locals("course").(*models.Course)
	var contents []models.CourseContent

	if err := database.DB.Where("course_id = ?", course.ID).
		Order("\"order\" ASC").
		Find(&contents).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  false,
			"message": "Failed to fetch content",
		})
	}

	return c.JSON(fiber.Map{
		"status": true,
		"data":   contents,
	})
}

// GetCourseReviews fetches reviews for a course
func GetCourseReviews(c *fiber.Ctx) error {
	courseID := c.Params("id")
	var reviews []models.Review

	if err := database.DB.Preload("User").
		Where("course_id = ?", courseID).
		Find(&reviews).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  false,
			"message": "Failed to fetch reviews",
		})
	}

	return c.JSON(fiber.Map{
		"status": true,
		"data":   reviews,
	})
}

// GetCertificateDetail gets certificate details
func GetCertificateDetail(c *fiber.Ctx) error {
	certID := c.Params("certificateId")
	var cert models.Certificate

	if err := database.DB.Preload("Course").
		Preload("User").
		First(&cert, "certificate_id = ?", certID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  false,
			"message": "Certificate not found",
		})
	}

	return c.JSON(fiber.Map{
		"status": true,
		"data":   cert,
	})
}
