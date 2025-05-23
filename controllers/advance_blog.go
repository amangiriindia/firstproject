package controllers

import (
	"firstproject/database"
	"firstproject/models"
	"firstproject/utils"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"time"
)

// GetPopularBlogs returns most viewed blogs
func GetPopularBlogs(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 10)
	days := c.QueryInt("days", 30) // Popular in last 30 days

	var blogs []models.Blog

	// Calculate date threshold
	threshold := time.Now().AddDate(0, 0, -days)

	if err := database.DB.Where("created_at >= ? AND status = ?", threshold, "published").
		Order("view_count DESC").
		Limit(limit).
		Preload("Category").
		Preload("Author", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "username")
		}).
		Find(&blogs).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch popular blogs"})
	}

	// Add comment counts
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

	return c.JSON(blogResponses)
}

// GetRecentBlogs returns recently published blogs
func GetRecentBlogs(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 10)

	var blogs []models.Blog

	if err := database.DB.Where("status = ?", "published").
		Order("created_at DESC").
		Limit(limit).
		Preload("Category").
		Preload("Author", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "username")
		}).
		Find(&blogs).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch recent blogs"})
	}

	return c.JSON(blogs)
}

// GetFeaturedBlogs returns featured blogs (high view count + recent)
func GetFeaturedBlogs(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 5)

	var blogs []models.Blog

	// Get blogs with high engagement (views + comments) from last 60 days
	threshold := time.Now().AddDate(0, 0, -60)

	if err := database.DB.Raw(`
		SELECT b.*, 
		       (b.view_count + COALESCE(comment_counts.count, 0) * 5) as engagement_score
		FROM blogs b
		LEFT JOIN (
			SELECT blog_id, COUNT(*) as count 
			FROM comments 
			GROUP BY blog_id
		) comment_counts ON b.id = comment_counts.blog_id
		WHERE b.status = ? AND b.created_at >= ?
		ORDER BY engagement_score DESC
		LIMIT ?
	`, "published", threshold, limit).Scan(&blogs).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch featured blogs"})
	}

	// Load relationships
	for i := range blogs {
		database.DB.Preload("Category").
			Preload("Author", func(db *gorm.DB) *gorm.DB {
				return db.Select("id", "username")
			}).
			First(&blogs[i], blogs[i].ID)
	}

	return c.JSON(blogs)
}

// GetBlogsByAuthor returns all blogs by a specific author
func GetBlogsByAuthor(c *fiber.Ctx) error {
	authorID := c.Params("id")
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	status := c.Query("status", "published")

	var blogs []models.Blog
	var total int64

	query := database.DB.Where("author_id = ?", authorID)
	if status != "all" {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)

	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).
		Order("created_at DESC").
		Preload("Category").
		Preload("Author", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "username")
		}).
		Find(&blogs).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch author's blogs"})
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	return c.JSON(fiber.Map{
		"blogs":       blogs,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": totalPages,
	})
}

// GetRelatedBlogs returns blogs related to a specific blog (same category or similar keywords)
func GetRelatedBlogs(c *fiber.Ctx) error {
	blogID := c.Params("id")
	limit := c.QueryInt("limit", 5)

	// Get the original blog
	var originalBlog models.Blog
	if err := database.DB.First(&originalBlog, blogID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Blog not found"})
	}

	var relatedBlogs []models.Blog

	// First try to find blogs in the same category
	if err := database.DB.Where("category_id = ? AND id != ? AND status = ?",
		originalBlog.CategoryID, originalBlog.ID, "published").
		Order("created_at DESC").
		Limit(limit).
		Preload("Category").
		Preload("Author", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "username")
		}).
		Find(&relatedBlogs).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch related blogs"})
	}

	// If not enough blogs in same category, find by similar keywords
	if len(relatedBlogs) < limit && originalBlog.Keywords != "" {
		remainingLimit := limit - len(relatedBlogs)
		var keywordBlogs []models.Blog

		searchTerm := "%" + originalBlog.Keywords + "%"
		if err := database.DB.Where("keywords ILIKE ? AND id != ? AND category_id != ? AND status = ?",
			searchTerm, originalBlog.ID, originalBlog.CategoryID, "published").
			Order("created_at DESC").
			Limit(remainingLimit).
			Preload("Category").
			Preload("Author", func(db *gorm.DB) *gorm.DB {
				return db.Select("id", "username")
			}).
			Find(&keywordBlogs).Error; err == nil {
			relatedBlogs = append(relatedBlogs, keywordBlogs...)
		}
	}

	return c.JSON(relatedBlogs)
}

// GetBlogStats returns statistics for a blog (views, comments, etc.)
func GetBlogStats(c *fiber.Ctx) error {
	blogID := c.Params("id")

	// Verify blog exists and user owns it
	var blog models.Blog
	if err := database.DB.First(&blog, blogID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Blog not found"})
	}

	user := c.Locals("user").(*models.User)
	if blog.AuthorID != user.ID {
		return c.Status(403).JSON(fiber.Map{"error": "You are not authorized to view these stats"})
	}

	// Get comment count
	var commentCount int64
	database.DB.Model(&models.Comment{}).Where("blog_id = ?", blog.ID).Count(&commentCount)

	// Get comment count by day for last 30 days
	var dailyComments []struct {
		Date  string `json:"date"`
		Count int    `json:"count"`
	}

	database.DB.Raw(`
		SELECT DATE(created_at) as date, COUNT(*) as count
		FROM comments 
		WHERE blog_id = ? AND created_at >= NOW() - INTERVAL '30 days'
		GROUP BY DATE(created_at)
		ORDER BY date
	`, blog.ID).Scan(&dailyComments)

	stats := fiber.Map{
		"blog_id":        blog.ID,
		"view_count":     blog.ViewCount,
		"comment_count":  commentCount,
		"created_at":     blog.CreatedAt,
		"updated_at":     blog.UpdatedAt,
		"status":         blog.Status,
		"daily_comments": dailyComments,
		"reading_time":   utils.CalculateReadingTime(blog.Content),
	}

	return c.JSON(stats)
}

// GetDashboardStats returns overall dashboard statistics for authenticated user
func GetDashboardStats(c *fiber.Ctx) error {
	user := c.Locals("user").(*models.User)

	// Total blogs by user
	var totalBlogs int64
	database.DB.Model(&models.Blog{}).Where("author_id = ?", user.ID).Count(&totalBlogs)

	// Total views on user's blogs
	var totalViews int64
	database.DB.Model(&models.Blog{}).Where("author_id = ?", user.ID).
		Select("COALESCE(SUM(view_count), 0)").Scan(&totalViews)

	// Total comments on user's blogs
	var totalComments int64
	database.DB.Raw(`
		SELECT COUNT(*) 
		FROM comments c 
		JOIN blogs b ON c.blog_id = b.id 
		WHERE b.author_id = ?
	`, user.ID).Scan(&totalComments)

	// Blogs by status
	var blogsByStatus []struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}
	database.DB.Model(&models.Blog{}).
		Select("status, COUNT(*) as count").
		Where("author_id = ?", user.ID).
		Group("status").
		Scan(&blogsByStatus)

	// Recent activity (last 7 days)
	var recentViews int64
	var recentComments int64

	sevenDaysAgo := time.Now().AddDate(0, 0, -7)

	database.DB.Model(&models.Blog{}).
		Where("author_id = ? AND updated_at >= ?", user.ID, sevenDaysAgo).
		Select("COALESCE(SUM(view_count), 0)").Scan(&recentViews)

	database.DB.Raw(`
		SELECT COUNT(*) 
		FROM comments c 
		JOIN blogs b ON c.blog_id = b.id 
		WHERE b.author_id = ? AND c.created_at >= ?
	`, user.ID, sevenDaysAgo).Scan(&recentComments)

	stats := fiber.Map{
		"total_blogs":     totalBlogs,
		"total_views":     totalViews,
		"total_comments":  totalComments,
		"blogs_by_status": blogsByStatus,
		"recent_views":    recentViews,
		"recent_comments": recentComments,
	}

	return c.JSON(stats)
}

// ToggleBlogStatus toggles blog between published/draft
func ToggleBlogStatus(c *fiber.Ctx) error {
	blogID := c.Params("id")
	var blog models.Blog

	if err := database.DB.First(&blog, blogID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Blog not found"})
	}

	user := c.Locals("user").(*models.User)
	if blog.AuthorID != user.ID {
		return c.Status(403).JSON(fiber.Map{"error": "You are not authorized to modify this blog"})
	}

	// Toggle status
	if blog.Status == "published" {
		blog.Status = "draft"
	} else {
		blog.Status = "published"
	}

	if err := database.DB.Save(&blog).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update blog status"})
	}

	return c.JSON(fiber.Map{
		"message": "Blog status updated successfully",
		"status":  blog.Status,
	})
}
