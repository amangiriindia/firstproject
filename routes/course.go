package routes

import (
	"firstproject/controllers"
	"firstproject/middleware"
	"firstproject/validators"
	"github.com/gofiber/fiber/v2"
)

func CourseRoutes(app *fiber.App) {
	// Course routes group
	course := app.Group("/api/courses")

	// Public routes - no authentication required
	course.Get("/", controllers.GetCourses)   // GET /api/courses - List all published courses with filtering/pagination
	course.Get("/:id", controllers.GetCourse) // GET /api/courses/:id - Get single course details

	// Authenticated routes - require valid JWT token
	authenticated := course.Use(middleware.JWTMiddleware)

	// Course management (for instructors/authors)
	authenticated.Post("/",
		validators.CreateCourseValidator,
		controllers.CreateCourse) // POST /api/courses - Create new course

	authenticated.Put("/:id",
		middleware.AuthorMiddleware,
		validators.UpdateCourseValidator,
		controllers.UpdateCourse) // PUT /api/courses/:id - Update course (author only)

	authenticated.Delete("/:id",
		middleware.AuthorMiddleware,
		controllers.DeleteCourse) // DELETE /api/courses/:id - Delete course (author only)

	authenticated.Put("/:id/publish",
		middleware.AuthorMiddleware,
		controllers.PublishCourse) // PUT /api/courses/:id/publish - Publish course (author only)

	// Course content management (for instructors/authors)
	authenticated.Post("/:id/content",
		middleware.AuthorMiddleware,
		validators.CreateContentValidator,
		controllers.AddCourseContent) // POST /api/courses/:id/content - Add content to course

	// Student enrollment and progress
	authenticated.Post("/:id/enroll", controllers.EnrollCourse) // POST /api/courses/:id/enroll - Enroll in course

	// Reviews and ratings
	authenticated.Post("/:id/reviews",
		validators.ReviewValidator,
		controllers.AddReview) // POST /api/courses/:id/reviews - Add course review

	// Certificates
	certificates := app.Group("/api/certificates", middleware.JWTMiddleware)
	certificates.Get("/", controllers.GetCertificates) // GET /api/certificates - Get user's certificates
}
