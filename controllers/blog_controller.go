package controllers

import (
	"firstproject/database"
	"firstproject/models"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func GetAllBlogs(c *fiber.Ctx) error {
	var blogs []models.Blog
	if err := database.DB.Preload("Category").Preload("Author", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "username")
	}).Find(&blogs).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch blogs"})
	}
	return c.JSON(blogs)
}

func GetBlogByID(c *fiber.Ctx) error {
	id := c.Params("id")
	var blog models.Blog
	if err := database.DB.Preload("Category").Preload("Author", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "username")
	}).First(&blog, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Blog not found"})
	}
	return c.JSON(blog)
}

func GetBlogsByCategory(c *fiber.Ctx) error {
	categoryID := c.Params("id")
	var blogs []models.Blog
	if err := database.DB.Where("category_id = ?", categoryID).Preload("Category").Preload("Author", func(db *gorm.DB) *gorm.DB {
		return db.Select("id", "username")
	}).Find(&blogs).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch blogs"})
	}
	return c.JSON(blogs)
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

	if err := database.DB.Delete(&blog).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete blog"})
	}

	return c.SendStatus(204)
}

func GetAllCategories(c *fiber.Ctx) error {
	var categories []models.Category
	if err := database.DB.Find(&categories).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch categories"})
	}
	return c.JSON(categories)
}
