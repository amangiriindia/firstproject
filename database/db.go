package database

import (
	"firstproject/config"
	"firstproject/models"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
)

var DB *gorm.DB

func ConnectDB() {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		config.GetEnv("DB_HOST"),
		config.GetEnv("DB_USER"),
		config.GetEnv("DB_PASSWORD"),
		config.GetEnv("DB_NAME"),
		config.GetEnv("DB_PORT"),
	)

	var err error
	// Fix: Assign to the global DB variable, not a local variable
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to PostgreSQL database:", err)
	}

	log.Println("Database connected successfully!")

	// Auto-migrate models
	err = DB.AutoMigrate(
		&models.User{},
		&models.UserProfile{},
		&models.Blog{},
		&models.Category{},
		&models.Comment{},
		// Note: Don't migrate BlogInput as it's just an input struct, not a database model
	)

	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	log.Println("Database migration completed successfully!")
}
