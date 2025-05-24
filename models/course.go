package models

import (
	"github.com/lib/pq"
	"time"
)

// Course represents a course in the platform
type Course struct {
	ID            uint            `gorm:"primaryKey" json:"id"`
	Title         string          `gorm:"not null;size:255" json:"title"`
	Description   string          `gorm:"type:text" json:"description"`
	FeaturedImage string          `gorm:"size:500" json:"featured_image"`
	Price         float64         `gorm:"default:0" json:"price"` // 0 for free
	Currency      string          `gorm:"default:USD;size:3" json:"currency"`
	Level         string          `gorm:"default:beginner" json:"level"` // beginner, intermediate, advanced
	Duration      int             `json:"duration"`                      // in minutes
	Language      string          `gorm:"default:English;size:50" json:"language"`
	Category      string          `gorm:"size:100" json:"category"`
	Tags          pq.StringArray  `gorm:"type:text[]" json:"tags"`
	IsPublished   bool            `gorm:"default:false" json:"is_published"`
	AuthorID      uint            `gorm:"not null" json:"author_id"`
	Author        User            `gorm:"foreignKey:AuthorID" json:"author,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
	Contents      []CourseContent `gorm:"foreignKey:CourseID;constraint:OnDelete:CASCADE" json:"contents,omitempty"`
	Enrollments   []Enrollment    `gorm:"foreignKey:CourseID;constraint:OnDelete:CASCADE" json:"enrollments,omitempty"`
}

// CourseContent represents different types of content within a course
type CourseContent struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	CourseID    uint      `gorm:"not null" json:"course_id"`
	Title       string    `gorm:"not null;size:255" json:"title"`
	Type        string    `gorm:"not null;size:50" json:"type"` // 'mcq', 'pdf', 'text', 'video', 'image', 'note', 'assignment'
	Data        string    `gorm:"type:json" json:"data"`        // JSON string for content specifics
	Duration    int       `json:"duration"`                     // in minutes for videos
	Order       int       `gorm:"not null" json:"order"`
	IsPreview   bool      `gorm:"default:false" json:"is_preview"` // Can be viewed without enrollment
	IsCompleted bool      `gorm:"default:false" json:"is_completed"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Enrollment represents a user's enrollment in a course
type Enrollment struct {
	ID                uint                   `gorm:"primaryKey" json:"id"`
	UserID            uint                   `gorm:"not null" json:"user_id"`
	CourseID          uint                   `gorm:"not null" json:"course_id"`
	User              User                   `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Course            Course                 `gorm:"foreignKey:CourseID" json:"course,omitempty"`
	EnrolledAt        time.Time              `json:"enrolled_at"`
	CompletedAt       *time.Time             `json:"completed_at,omitempty"`
	LastAccessedAt    *time.Time             `json:"last_accessed_at,omitempty"`
	Progress          int                    `gorm:"default:0" json:"progress"` // Percentage completed (0-100)
	CurrentContentID  *uint                  `json:"current_content_id,omitempty"`
	PaymentStatus     string                 `gorm:"default:pending" json:"payment_status"` // pending, completed, failed
	PaymentID         string                 `json:"payment_id,omitempty"`
	CompletedContents pq.Int64Array          `gorm:"type:integer[]" json:"completed_contents"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
	ContentProgress   []ContentProgress      `gorm:"foreignKey:EnrollmentID;constraint:OnDelete:CASCADE" json:"content_progress,omitempty"`
	Assignments       []AssignmentSubmission `gorm:"foreignKey:EnrollmentID;constraint:OnDelete:CASCADE" json:"assignments,omitempty"`
}

// ContentProgress tracks progress for individual content items
type ContentProgress struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	EnrollmentID uint       `gorm:"not null" json:"enrollment_id"`
	ContentID    uint       `gorm:"not null" json:"content_id"`
	IsCompleted  bool       `gorm:"default:false" json:"is_completed"`
	TimeSpent    int        `gorm:"default:0" json:"time_spent"`    // in seconds
	LastPosition int        `gorm:"default:0" json:"last_position"` // for videos
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// AssignmentSubmission represents assignment submissions
type AssignmentSubmission struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	EnrollmentID uint       `gorm:"not null" json:"enrollment_id"`
	ContentID    uint       `gorm:"not null" json:"content_id"`
	Submission   string     `gorm:"type:text" json:"submission"`
	FileURL      string     `json:"file_url,omitempty"`
	Grade        *float64   `json:"grade,omitempty"`
	Feedback     string     `gorm:"type:text" json:"feedback,omitempty"`
	SubmittedAt  time.Time  `json:"submitted_at"`
	GradedAt     *time.Time `json:"graded_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// Certificate represents completion certificates
type Certificate struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	UserID        uint       `gorm:"not null" json:"user_id"`
	CourseID      uint       `gorm:"not null" json:"course_id"`
	User          User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Course        Course     `gorm:"foreignKey:CourseID" json:"course,omitempty"`
	CertificateID string     `gorm:"unique;not null" json:"certificate_id"`
	IssuedAt      time.Time  `json:"issued_at"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// Review represents course reviews
type Review struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null" json:"user_id"`
	CourseID  uint      `gorm:"not null" json:"course_id"`
	User      User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Course    Course    `gorm:"foreignKey:CourseID" json:"course,omitempty"`
	Rating    int       `gorm:"check:rating >= 1 AND rating <= 5" json:"rating"`
	Comment   string    `gorm:"type:text" json:"comment"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Quiz represents quizzes/MCQs
type Quiz struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	ContentID     uint           `gorm:"not null" json:"content_id"`
	Question      string         `gorm:"type:text;not null" json:"question"`
	Options       pq.StringArray `gorm:"type:text[]" json:"options"`
	CorrectAnswer int            `gorm:"not null" json:"correct_answer"` // index of correct option
	Explanation   string         `gorm:"type:text" json:"explanation"`
	Points        int            `gorm:"default:1" json:"points"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}
