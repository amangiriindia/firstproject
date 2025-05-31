package routes

import (
	"firstproject/controllers"
	"firstproject/validators"

	"github.com/gofiber/fiber/v2"
)

func AuthRoutes(app *fiber.App) {
	auth := app.Group("/auth")

	// Public authentication routes
	auth.Post("/register", validators.RegisterValidator, controllers.Register)
	auth.Post("/login", validators.LoginValidator, controllers.Login)
	auth.Post("/forgot-password", validators.ForgotPasswordValidator, controllers.ForgotPassword)
	auth.Post("/reset-password", validators.ResetPasswordValidator, controllers.ResetPassword)
	auth.Post("/verify-email", controllers.VerifyEmail)
	auth.Post("/resend-verification", controllers.ResendVerification)
}
