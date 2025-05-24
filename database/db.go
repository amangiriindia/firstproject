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

		// Course platform models
		&models.Course{},
		&models.CourseContent{},
		&models.Enrollment{},
		&models.ContentProgress{},
		&models.AssignmentSubmission{},
		&models.Certificate{},
		&models.Review{},
		&models.Quiz{},
	)

	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	log.Println("Database migration completed successfully!")
}
