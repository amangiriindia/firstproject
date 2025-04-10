package middleware

import (
	"firstproject/models"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// var jwtSecret = []byte("asdfasqsdfgsdasdfasdfawqe") // Replace with your actual secret key

// GenerateJWT generates a JWT token for the user

// Helper function to generate JWT
func GenerateJWT(user models.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	}

	// Create the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtSecret := []byte(os.Getenv("JWT_SECRET"))

	// Sign the token with the secret key
	return token.SignedString(jwtSecret)
}

// JWTMiddleware is a middleware to check for valid JWT token in the request
func JWTMiddleware(c *fiber.Ctx) error {
	// Get the token from the Authorization header
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  false,
			"message": "Missing or invalid Authorization header",
		})
	}

	// The token should be prefixed with "Bearer "
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  false,
			"message": "Invalid Authorization header format",
		})
	}

	// Extract the token part
	tokenString := authHeader[len("Bearer "):]

	// Parse and validate the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Check if the token method is valid
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		jwtSecret := []byte(os.Getenv("JWT_SECRET"))
		return jwtSecret, nil
	})

	// If there's an error parsing the token
	if err != nil || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  false,
			"message": "Invalid or expired token",
		})
	}

	// Extract user ID from the token claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || claims["userId"] == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  false,
			"message": "Invalid token payload",
		})
	}

	// Set the user ID in the request context
	userID := claims["userId"].(float64) // JWT claims are typically stored as `float64`, so cast it
	c.Locals("userId", uint(userID))     // Store userID in context as uint

	// If valid, continue to the next handler
	return c.Next()
}

func JsonResponse(c *fiber.Ctx, statusCode int, status bool, message string, data interface{}) error {
	return c.Status(statusCode).JSON(fiber.Map{
		"status":  status,
		"message": message,
		"data":    data,
	})
}

func ValidationErrorResponse(c *fiber.Ctx, errors map[string]string) error {
	return JsonResponse(c, fiber.StatusUnprocessableEntity, false, "Validation failed!", errors)
}
