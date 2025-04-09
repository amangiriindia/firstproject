package routes

import (
	"firstproject/controllers"
	"github.com/gofiber/fiber/v2"
)

// AuthRoutes groups all authentication routes under /auth
func AuthRoutes(app *fiber.App) {
	auth := app.Group("/auth")

	// Basic authentication
	auth.Post("/register", controllers.Register)
	auth.Post("/login", controllers.Login)
	auth.Post("/refresh", controllers.RefreshToken)
	auth.Post("/logout", controllers.Logout)

	// Password reset
	auth.Post("/forgot-password", controllers.ForgotPassword)
	auth.Post("/reset-password", controllers.ResetPassword)

	// Email verification
	auth.Post("/verify-email", controllers.VerifyEmail)
	auth.Post("/resend-verification", controllers.ResendVerification)

	// Social auth (if implemented)
	// auth.Get("/:provider", controllers.InitiateSocialAuth) // e.g., /auth/google
	// auth.Get("/:provider/callback", controllers.SocialAuthCallback)

	// Profile management
	auth.Get("/me", controllers.GetCurrentUser)
	auth.Put("/me", controllers.UpdateProfile)
}
