package models

import (
	"github.com/lib/pq"
	"time"
)

type User struct {
	ID                uint        `gorm:"primaryKey" json:"id"`
	Username          string      `gorm:"unique;not null" json:"username"`
	Email             string      `gorm:"unique;not null" json:"email"`
	Mobile            string      `gorm:"unique;not null" json:"mobile"`
	Password          string      `gorm:"-" json:"password,omitempty"`
	PasswordHash      string      `gorm:"not null" json:"-"`
	FirstName         string      `json:"first_name"`
	LastName          string      `json:"last_name"`
	AvatarURL         string      `json:"avatar_url"`
	Bio               string      `json:"bio"`
	Role              string      `gorm:"default:user" json:"role"`
	IsVerified        bool        `gorm:"default:false" json:"is_verified"`
	VerificationToken string      `json:"-"`
	ResetToken        string      `json:"-"`
	ResetTokenExpires time.Time   `json:"-"`
	CreatedAt         time.Time   `json:"created_at"`
	UpdatedAt         time.Time   `json:"updated_at"`
	Profile           UserProfile `gorm:"foreignKey:UserID" json:"profile"`
}

type UserProfile struct {
	ID          uint           `gorm:"primaryKey" json:"-"`
	UserID      uint           `json:"-"`
	Skills      pq.StringArray `gorm:"type:text[]" json:"skills"`
	Interests   pq.StringArray `gorm:"type:text[]" json:"interests"`
	GithubURL   string         `json:"github_url"`
	LinkedinURL string         `json:"linkedin_url"`
	TwitterURL  string         `json:"twitter_url"`
	WebsiteURL  string         `json:"website_url"`
	Education   string         `json:"education"`
	Experience  string         `json:"experience"`
	CreatedAt   time.Time      `json:"-"`
	UpdatedAt   time.Time      `json:"-"`
}
