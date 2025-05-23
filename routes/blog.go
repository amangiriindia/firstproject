package routes

import (
	"firstproject/controllers"
	"firstproject/middleware"
	"firstproject/validators"
	"github.com/gofiber/fiber/v2"
)

func BlogRoutes(app *fiber.App) {
	// Public routes for blogs
	app.Get("/blogs", validators.SearchParamsValidator, controllers.GetAllBlogs)
	app.Get("/blogs/search", validators.SearchParamsValidator, controllers.SearchBlogs)
	app.Get("/blogs/:id", controllers.GetBlogByID)
	app.Get("/categories/:id/blogs", controllers.GetBlogsByCategory)

	// Public routes for categories
	app.Get("/categories", controllers.GetAllCategories)

	// Public routes for comments (read only)
	app.Get("/blogs/:id/comments", controllers.GetBlogComments)

	// Protected routes with JWT middleware for blogs
	blogGroup := app.Group("/blogs", middleware.JWTMiddleware)
	blogGroup.Post("/", validators.CreateBlogValidator, controllers.CreateBlog)
	blogGroup.Put("/:id", validators.CreateBlogValidator, controllers.UpdateBlog)
	blogGroup.Delete("/:id", controllers.DeleteBlog)

	// Protected routes with JWT middleware for categories
	categoryGroup := app.Group("/categories", middleware.JWTMiddleware)
	categoryGroup.Post("/", validators.CreateCategoryValidator, controllers.CreateCategory)
	categoryGroup.Put("/:id", validators.CreateCategoryValidator, controllers.UpdateCategory)
	categoryGroup.Delete("/:id", controllers.DeleteCategory)

	// Protected routes with JWT middleware for comments
	commentGroup := app.Group("/comments", middleware.JWTMiddleware)
	commentGroup.Post("/", validators.CreateCommentValidator, controllers.CreateComment)
	commentGroup.Put("/:id", validators.CreateCommentValidator, controllers.UpdateComment)
	commentGroup.Delete("/:id", controllers.DeleteComment)

	// Additional public routes for enhanced features
	app.Get("/blogs/popular", controllers.GetPopularBlogs)
	app.Get("/blogs/recent", controllers.GetRecentBlogs)
	app.Get("/blogs/featured", controllers.GetFeaturedBlogs)
	app.Get("/blogs/:id/related", controllers.GetRelatedBlogs)
	app.Get("/authors/:id/blogs", controllers.GetBlogsByAuthor)

	// Protected dashboard and analytics routes
	dashboardGroup := app.Group("/dashboard", middleware.JWTMiddleware)
	dashboardGroup.Get("/stats", controllers.GetDashboardStats)
	dashboardGroup.Get("/blogs/:id/stats", controllers.GetBlogStats)
	dashboardGroup.Patch("/blogs/:id/toggle-status", controllers.ToggleBlogStatus)
}
