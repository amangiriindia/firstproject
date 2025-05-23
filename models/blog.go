package models

import "gorm.io/gorm"

type Blog struct {
	gorm.Model
	Title      string   `json:"title"`
	Content    string   `json:"content"`
	AuthorID   uint     `json:"-"`
	Author     User     `json:"author" gorm:"foreignKey:AuthorID"`
	CategoryID uint     `json:"-"`
	Category   Category `json:"category" gorm:"foreignKey:CategoryID"`
	ImageURL   string   `json:"image_url"`
}

type Category struct {
	gorm.Model
	Name        string `json:"name"`
	Description string `json:"description"`
}

type BlogInput struct {
	Title      string `json:"title"`
	Content    string `json:"content"`
	CategoryID uint   `json:"category_id"`
	ImageURL   string `json:"image_url"`
}
