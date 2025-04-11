package routes

import (
	"firstproject/controllers"
	"firstproject/middleware"

	"github.com/gofiber/fiber/v2"
)

func UserRoutes(app *fiber.App) {
	// Grouping protected routes using the JWT middleware
	userGroup := app.Group("/user", middleware.JWTMiddleware)
	userGroup.Get("/profile", controllers.GetProfile)
	userGroup.Put("/profile", controllers.UpdateProfile)
}
