package routes

import (
	"firstproject/controllers"
	"github.com/gofiber/fiber/v2"
	jwtware "github.com/gofiber/jwt/v3"
	"os"
)

func AuthRoutes(app *fiber.App) {
	auth := app.Group("/auth")

	// Public authentication routes
	auth.Post("/register", controllers.Register)
	auth.Post("/login", controllers.Login)
	auth.Post("/refresh", controllers.RefreshToken)
	auth.Post("/logout", controllers.Logout)
	auth.Post("/forgot-password", controllers.ForgotPassword)
	auth.Post("/reset-password", controllers.ResetPassword)
	auth.Post("/verify-email", controllers.VerifyEmail)
	auth.Post("/resend-verification", controllers.ResendVerification)

	// Secure routes: apply the JWT middleware
	auth.Use(jwtware.New(jwtware.Config{
		SigningKey: []byte(os.Getenv("JWT_SECRET")),
		ContextKey: "user", // this key is used by your handlers (e.g., c.Locals("user"))
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized: Invalid or missing token",
			})
		},
	}))

	// These routes now require a valid JWT token in the Authorization header
	auth.Get("/me", controllers.GetCurrentUser)
	auth.Put("/me", controllers.UpdateProfile)
}
