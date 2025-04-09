package main

import (
	"firstproject/config"
	"firstproject/database"
	"firstproject/routes"
	"github.com/gofiber/fiber/v2"
)

func main() {
	config.LoadEnv()
	database.ConnectDB()

	app := fiber.New()

	routes.AuthRoutes(app)

	app.Listen(":3000")
}
