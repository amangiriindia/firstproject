package controllers

import (
	"firstproject/database"
	"firstproject/models"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"math"
)

// Blog Controllers
func GetAllBlogs(c *fiber.Ctx) error {
	params := c.Locals("searchParams").(models.BlogSearchParams)

	var blogs []models.Blog
	var total int64

	query := database.DB.Model(&models.Blog{})

	// Apply filters
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}
	if params.CategoryID > 0 {
		query = query.Where("category_id = ?", params.CategoryID)
	}
	if params.AuthorID > 0 {
		query = query.Where("author_id = ?", params.AuthorID)
	}
	if params.Query != "" {
		searchTerm := "%" + params.Query + "%"
		query = query.Where("title ILIKE ? OR content ILIKE ? OR keywords ILIKE ?",
			searchTerm, searchTerm, searchTerm)
	}

	// Count total records
	query.Count(&total)

	// Apply sorting
	orderClause := params.SortBy + " " + params.Order
	query = query.Order(orderClause)

	// Apply pagination
	offset := (params.Page - 1) * params.Limit
	query = query.Offset(offset).Limit(params.Limit)

	// Execute query with preloads
	if err := query.Preload("Category").Preload("Author", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "username")
	}).Find(&blogs).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch blogs"})
	}

	// Convert to response format with comment counts
	var blogResponses []models.BlogResponse
	for _, blog := range blogs {
		var commentCount int64
		database.DB.Model(&models.Comment{}).Where("blog_id = ?", blog.ID).Count(&commentCount)

		blogResponse := models.BlogResponse{
			Blog:         blog,
			CommentCount: int(commentCount),
		}
		blogResponses = append(blogResponses, blogResponse)
	}

	totalPages := int(math.Ceil(float64(total) / float64(params.Limit)))

	response := models.PaginatedBlogResponse{
		Blogs:      blogResponses,
		Total:      total,
		Page:       params.Page,
		Limit:      params.Limit,
		TotalPages: totalPages,
	}

	return c.JSON(response)
}

func GetBlogByID(c *fiber.Ctx) error {
	id := c.Params("id")
	var blog models.Blog

	if err := database.DB.Preload("Category").Preload("Author", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "username")
	}).First(&blog, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Blog not found"})
	}

	// Increment view count
	database.DB.Model(&blog).UpdateColumn("view_count", gorm.Expr("view_count + ?", 1))

	// Get comment count
	var commentCount int64
	database.DB.Model(&models.Comment{}).Where("blog_id = ?", blog.ID).Count(&commentCount)

	response := models.BlogResponse{
		Blog:         blog,
		CommentCount: int(commentCount),
	}

	return c.JSON(response)
}

func GetBlogsByCategory(c *fiber.Ctx) error {
	categoryID := c.Params("id")
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)

	var blogs []models.Blog
	var total int64

	query := database.DB.Where("category_id = ? AND status = ?", categoryID, "published")
	query.Count(&total)

	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).
		Preload("Category").
		Preload("Author", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "username")
		}).
		Order("created_at DESC").
		Find(&blogs).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch blogs"})
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return c.JSON(fiber.Map{
		"blogs":       blogs,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": totalPages,
	})
}

func CreateBlog(c *fiber.Ctx) error {
	input := c.Locals("blogInput").(models.BlogInput)
	user := c.Locals("user").(*models.User)

	var category models.Category
	if err := database.DB.First(&category, input.CategoryID).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid category"})
	}

	blog := models.Blog{
		Title:      input.Title,
		Content:    input.Content,
		AuthorID:   user.ID,
		CategoryID: input.CategoryID,
		ImageURL:   input.ImageURL,
		Keywords:   input.Keywords,
		Status:     input.Status,
	}

	if err := database.DB.Create(&blog).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create blog"})
	}

	if err := database.DB.Preload("Category").Preload("Author", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "username")
	}).First(&blog, blog.ID).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch created blog"})
	}

	return c.Status(201).JSON(blog)
}

func UpdateBlog(c *fiber.Ctx) error {
	id := c.Params("id")
	var blog models.Blog
	if err := database.DB.First(&blog, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Blog not found"})
	}

	user := c.Locals("user").(*models.User)
	if blog.AuthorID != user.ID {
		return c.Status(403).JSON(fiber.Map{"error": "You are not authorized to update this blog"})
	}

	input := c.Locals("blogInput").(models.BlogInput)

	var category models.Category
	if err := database.DB.First(&category, input.CategoryID).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid category"})
	}

	blog.Title = input.Title
	blog.Content = input.Content
	blog.CategoryID = input.CategoryID
	blog.ImageURL = input.ImageURL
	blog.Keywords = input.Keywords
	blog.Status = input.Status

	if err := database.DB.Save(&blog).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update blog"})
	}

	if err := database.DB.Preload("Category").Preload("Author", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "username")
	}).First(&blog, blog.ID).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch updated blog"})
	}

	return c.JSON(blog)
}

func DeleteBlog(c *fiber.Ctx) error {
	id := c.Params("id")
	var blog models.Blog
	if err := database.DB.First(&blog, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Blog not found"})
	}

	user := c.Locals("user").(*models.User)
	if blog.AuthorID != user.ID {
		return c.Status(403).JSON(fiber.Map{"error": "You are not authorized to delete this blog"})
	}

	// Delete associated comments first
	if err := database.DB.Where("blog_id = ?", blog.ID).Delete(&models.Comment{}).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete associated comments"})
	}

	if err := database.DB.Delete(&blog).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete blog"})
	}

	return c.SendStatus(204)
}

// Category Controllers
func GetAllCategories(c *fiber.Ctx) error {
	var categories []models.Category
	query := database.DB.Model(&models.Category{})

	// Only active categories for public endpoint
	if c.Query("include_inactive") != "true" {
		query = query.Where("is_active = ?", true)
	}

	if err := query.Order("name ASC").Find(&categories).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch categories"})
	}
	return c.JSON(categories)
}

func CreateCategory(c *fiber.Ctx) error {
	input := c.Locals("categoryInput").(models.CategoryInput)

	// Check if category name already exists
	var existingCategory models.Category
	if err := database.DB.Where("LOWER(name) = LOWER(?)", input.Name).First(&existingCategory).Error; err == nil {
		return c.Status(400).JSON(fiber.Map{"error": "Category name already exists"})
	}

	category := models.Category{
		Name:        input.Name,
		Description: input.Description,
		Color:       input.Color,
		IsActive:    input.IsActive,
	}

	if err := database.DB.Create(&category).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create category"})
	}

	return c.Status(201).JSON(category)
}

func UpdateCategory(c *fiber.Ctx) error {
	id := c.Params("id")
	var category models.Category
	if err := database.DB.First(&category, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Category not found"})
	}

	input := c.Locals("categoryInput").(models.CategoryInput)

	// Check if category name already exists (excluding current category)
	var existingCategory models.Category
	if err := database.DB.Where("LOWER(name) = LOWER(?) AND id != ?", input.Name, category.ID).First(&existingCategory).Error; err == nil {
		return c.Status(400).JSON(fiber.Map{"error": "Category name already exists"})
	}

	category.Name = input.Name
	category.Description = input.Description
	category.Color = input.Color
	category.IsActive = input.IsActive

	if err := database.DB.Save(&category).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update category"})
	}

	return c.JSON(category)
}

func DeleteCategory(c *fiber.Ctx) error {
	id := c.Params("id")
	var category models.Category
	if err := database.DB.First(&category, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Category not found"})
	}

	// Check if category has blogs
	var blogCount int64
	database.DB.Model(&models.Blog{}).Where("category_id = ?", category.ID).Count(&blogCount)
	if blogCount > 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Cannot delete category with existing blogs"})
	}

	if err := database.DB.Delete(&category).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete category"})
	}

	return c.SendStatus(204)
}

// Comment Controllers
func GetBlogComments(c *fiber.Ctx) error {
	blogID := c.Params("id")
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)

	// Verify blog exists
	var blog models.Blog
	if err := database.DB.First(&blog, blogID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Blog not found"})
	}

	var comments []models.Comment
	var total int64

	query := database.DB.Where("blog_id = ? AND parent_id IS NULL", blogID)
	query.Count(&total)

	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).
		Preload("Author", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "username")
		}).
		Preload("Replies", func(db *gorm.DB) *gorm.DB {
			return db.Preload("Author", func(db *gorm.DB) *gorm.DB {
				return db.Select("id", "username")
			}).Order("created_at ASC")
		}).
		Order("created_at DESC").
		Find(&comments).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch comments"})
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return c.JSON(fiber.Map{
		"comments":    comments,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": totalPages,
	})
}

func CreateComment(c *fiber.Ctx) error {
	input := c.Locals("commentInput").(models.CommentInput)
	user := c.Locals("user").(*models.User)

	// Verify blog exists
	var blog models.Blog
	if err := database.DB.First(&blog, input.BlogID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Blog not found"})
	}

	// If it's a reply, verify parent comment exists
	if input.ParentID != nil {
		var parentComment models.Comment
		if err := database.DB.Where("id = ? AND blog_id = ?", *input.ParentID, input.BlogID).First(&parentComment).Error; err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "Parent comment not found"})
		}
	}

	comment := models.Comment{
		Content:    input.Content,
		BlogID:     input.BlogID,
		AuthorID:   user.ID,
		ParentID:   input.ParentID,
		IsApproved: true, // Auto-approve for now
	}

	if err := database.DB.Create(&comment).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create comment"})
	}

	if err := database.DB.Preload("Author", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "username")
	}).First(&comment, comment.ID).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch created comment"})
	}

	return c.Status(201).JSON(comment)
}

func UpdateComment(c *fiber.Ctx) error {
	id := c.Params("id")
	var comment models.Comment
	if err := database.DB.First(&comment, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Comment not found"})
	}

	user := c.Locals("user").(*models.User)
	if comment.AuthorID != user.ID {
		return c.Status(403).JSON(fiber.Map{"error": "You are not authorized to update this comment"})
	}

	input := c.Locals("commentInput").(models.CommentInput)
	comment.Content = input.Content

	if err := database.DB.Save(&comment).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update comment"})
	}

	if err := database.DB.Preload("Author", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "username")
	}).First(&comment, comment.ID).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch updated comment"})
	}

	return c.JSON(comment)
}

func DeleteComment(c *fiber.Ctx) error {
	id := c.Params("id")
	var comment models.Comment
	if err := database.DB.First(&comment, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Comment not found"})
	}

	user := c.Locals("user").(*models.User)
	if comment.AuthorID != user.ID {
		return c.Status(403).JSON(fiber.Map{"error": "You are not authorized to delete this comment"})
	}

	// Delete all replies first
	if err := database.DB.Where("parent_id = ?", comment.ID).Delete(&models.Comment{}).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete comment replies"})
	}

	if err := database.DB.Delete(&comment).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete comment"})
	}

	return c.SendStatus(204)
}

// Search function
func SearchBlogs(c *fiber.Ctx) error {
	return GetAllBlogs(c) // Uses the same logic with search parameters
}
