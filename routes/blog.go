package routes

import (
	"firstproject/controllers"
	"firstproject/middleware"
	"firstproject/validators"
	"github.com/gofiber/fiber/v2"
)

func BlogRoutes(app *fiber.App) {
	// Public routes
	app.Get("/blogs", controllers.GetAllBlogs)
	app.Get("/blogs/:id", controllers.GetBlogByID)
	app.Get("/categories/:id/blogs", controllers.GetBlogsByCategory)
	app.Get("/categories", controllers.GetAllCategories)

	// Protected routes with JWT middleware
	blogGroup := app.Group("/blogs", middleware.JWTMiddleware)
	blogGroup.Post("/", validators.CreateBlogValidator, controllers.CreateBlog)
	blogGroup.Put("/:id", validators.CreateBlogValidator, controllers.UpdateBlog)
	blogGroup.Delete("/:id", controllers.DeleteBlog)
}
