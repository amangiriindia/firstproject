package models

import (
	"gorm.io/gorm"
)

type Blog struct {
	gorm.Model
	Title      string    `json:"title"`
	Content    string    `json:"content"`
	AuthorID   uint      `json:"-"`
	Author     User      `json:"author" gorm:"foreignKey:AuthorID"`
	CategoryID uint      `json:"-"`
	Category   Category  `json:"category" gorm:"foreignKey:CategoryID"`
	ImageURL   string    `json:"image_url"`
	Keywords   string    `json:"keywords"` // Comma-separated keywords for search
	ViewCount  int       `json:"view_count" gorm:"default:0"`
	Comments   []Comment `json:"comments,omitempty" gorm:"foreignKey:BlogID"`
	Status     string    `json:"status" gorm:"default:'published'"` // published, draft, archived
}

// Category model
type Category struct {
	gorm.Model
	Name        string `json:"name" gorm:"unique"`
	Description string `json:"description"`
	Color       string `json:"color"` // For UI styling
	IsActive    bool   `json:"is_active" gorm:"default:true"`
}

// Comment model
type Comment struct {
	gorm.Model
	Content    string    `json:"content"`
	BlogID     uint      `json:"blog_id"`
	Blog       Blog      `json:"-" gorm:"foreignKey:BlogID"`
	AuthorID   uint      `json:"-"`
	Author     User      `json:"author" gorm:"foreignKey:AuthorID"`
	ParentID   *uint     `json:"parent_id"` // For nested comments/replies
	Parent     *Comment  `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Replies    []Comment `json:"replies,omitempty" gorm:"foreignKey:ParentID"`
	IsApproved bool      `json:"is_approved" gorm:"default:true"`
	LikeCount  int       `json:"like_count" gorm:"default:0"`
}

// Input models for validation
type BlogInput struct {
	Title      string `json:"title" validate:"required,min=3,max=200"`
	Content    string `json:"content" validate:"required,min=10"`
	CategoryID uint   `json:"category_id" validate:"required,gt=0"`
	ImageURL   string `json:"image_url"`
	Keywords   string `json:"keywords"` // Comma-separated keywords
	Status     string `json:"status" validate:"oneof=published draft archived"`
}

type CategoryInput struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Description string `json:"description" validate:"max=500"`
	Color       string `json:"color"`
	IsActive    bool   `json:"is_active"`
}

type CommentInput struct {
	Content  string `json:"content" validate:"required,min=1,max=1000"`
	BlogID   uint   `json:"blog_id" validate:"required,gt=0"`
	ParentID *uint  `json:"parent_id"` // Optional for replies
}

// Search and filter structures
type BlogSearchParams struct {
	Query      string `query:"q"`
	CategoryID uint   `query:"category_id"`
	AuthorID   uint   `query:"author_id"`
	Status     string `query:"status"`
	SortBy     string `query:"sort_by"` // created_at, view_count, title
	Order      string `query:"order"`   // asc, desc
	Page       int    `query:"page"`
	Limit      int    `query:"limit"`
}

// Response structures
type BlogResponse struct {
	Blog
	CommentCount int `json:"comment_count"`
}

type PaginatedBlogResponse struct {
	Blogs      []BlogResponse `json:"blogs"`
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	Limit      int            `json:"limit"`
	TotalPages int            `json:"total_pages"`
}

type CommentResponse struct {
	Comment
	ReplyCount int `json:"reply_count"`
}
