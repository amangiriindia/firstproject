// package routes

// import (
// 	"firstproject/controllers"
// 	"firstproject/middleware"
// 	"firstproject/validators"
// 	"github.com/gofiber/fiber/v2"
// )

// func CourseRoutes(app *fiber.App) {
// 	// Course routes group
// 	course := app.Group("/api/courses")

// 	// Public routes - no authentication required
// 	course.Get("/", controllers.GetCourses)   // GET /api/courses - List all published courses with filtering/pagination
// 	course.Get("/:id", controllers.GetCourse) // GET /api/courses/:id - Get single course details

// 	// Authenticated routes - require valid JWT token
// 	authenticated := course.Use(middleware.JWTMiddleware)

// 	// Course management (for instructors/authors)
// 	authenticated.Post("/",
// 		validators.CreateCourseValidator,
// 		controllers.CreateCourse) // POST /api/courses - Create new course

// 	authenticated.Put("/:id",
// 		middleware.AuthorMiddleware,
// 		validators.UpdateCourseValidator,
// 		controllers.UpdateCourse) // PUT /api/courses/:id - Update course (author only)

// 	authenticated.Delete("/:id",
// 		middleware.AuthorMiddleware,
// 		controllers.DeleteCourse) // DELETE /api/courses/:id - Delete course (author only)

// 	authenticated.Put("/:id/publish",
// 		middleware.AuthorMiddleware,
// 		controllers.PublishCourse) // PUT /api/courses/:id/publish - Publish course (author only)

// 	// Course content management (for instructors/authors)
// 	authenticated.Post("/:id/content",
// 		middleware.AuthorMiddleware,
// 		validators.CreateContentValidator,
// 		controllers.AddCourseContent) // POST /api/courses/:id/content - Add content to course

// 	// Student enrollment and progress

// 	// Reviews and ratings
// 	authenticated.Post("/:id/reviews",
// 		validators.ReviewValidator,
// 		controllers.AddReview) // POST /api/courses/:id/reviews - Add course review

// 	// Certificates
// 	certificates := app.Group("/api/certificates", middleware.JWTMiddleware)
// 	certificates.Get("/", controllers.GetCertificates) // GET /api/certificates - Get user's certificates
// }

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

	// Public routes
	course.Get("/", controllers.GetCourses)
	course.Get("/:id", controllers.GetCourse)

	// Authenticated routes
	authenticated := course.Use(middleware.JWTMiddleware)

	// Course management (Instructors)
	authenticated.Post("/", validators.CreateCourseValidator, controllers.CreateCourse)
	authenticated.Put("/:id", middleware.AuthorMiddleware, validators.UpdateCourseValidator, controllers.UpdateCourse)
	authenticated.Delete("/:id", middleware.AuthorMiddleware, controllers.DeleteCourse)
	authenticated.Put("/:id/publish", middleware.AuthorMiddleware, controllers.PublishCourse)
	authenticated.Post("/:id/enroll", controllers.EnrollCourse) // POST /api/courses/:id/enroll - Enroll in course

	// Course content management (Instructors)
	content := authenticated.Group("/:id/content", middleware.AuthorMiddleware)
	content.Post("/", validators.CreateContentValidator, controllers.AddCourseContent)
	content.Put("/:contentId", validators.CreateContentValidator, controllers.UpdateCourseContent)
	content.Delete("/:contentId", controllers.DeleteCourseContent)
	content.Get("/", controllers.GetAllCourseContent) // For managing content order

	// Student enrollment and progress
	// enrolled := app.Group("/api/enrolled-courses", middleware.JWTMiddleware, middleware.EnrollmentMiddleware)
	// enrolled.Get("/", controllers.GetEnrolledCourses)
	// enrolled.Get("/:courseId", controllers.GetEnrolledCourse)
	// enrolled.Get("/:courseId/progress", controllers.GetCourseProgress)
	// enrolled.Get("/:courseId/contents", controllers.GetCourseContent)
	// enrolled.Get("/:courseId/contents/:contentId", controllers.GetContentWithProgress)
	// enrolled.Put("/:courseId/contents/:contentId/progress", validators.UpdateProgressValidator, controllers.UpdateProgress)
	// enrolled.Post("/:courseId/complete", controllers.CompleteCourse)
	// enrolled.Get("/:courseId/next-content", controllers.GetNextContent)
	// enrolled.Get("/:courseId/resume", controllers.GetResumePosition)

	// Student enrollment and progress - corrected routing
	enrolled := app.Group("/api/enrolled-courses", middleware.JWTMiddleware)

	// Routes that don't need EnrollmentMiddleware (no courseId parameter)
	enrolled.Get("/", controllers.GetEnrolledCourses)

	// Routes that need EnrollmentMiddleware (have courseId parameter)
	enrolled.Get("/:courseId", middleware.EnrollmentMiddleware, controllers.GetEnrolledCourse)
	enrolled.Get("/:courseId/progress", middleware.EnrollmentMiddleware, controllers.GetCourseProgress)
	enrolled.Get("/:courseId/contents", middleware.EnrollmentMiddleware, controllers.GetCourseContent)
	enrolled.Get("/:courseId/contents/:contentId", middleware.EnrollmentMiddleware, controllers.GetContentWithProgress)
	enrolled.Put("/:courseId/contents/:contentId/progress", middleware.EnrollmentMiddleware, validators.UpdateProgressValidator, controllers.UpdateProgress)
	enrolled.Post("/:courseId/complete", middleware.EnrollmentMiddleware, controllers.CompleteCourse)
	enrolled.Get("/:courseId/next-content", middleware.EnrollmentMiddleware, controllers.GetNextContent)
	enrolled.Get("/:courseId/resume", middleware.EnrollmentMiddleware, controllers.GetResumePosition)

	// Reviews and ratings
	authenticated.Post("/:id/reviews", validators.ReviewValidator, controllers.AddReview)
	authenticated.Get("/:id/reviews", controllers.GetCourseReviews)

	// Certificates
	certificates := app.Group("/api/certificates", middleware.JWTMiddleware)
	certificates.Get("/", controllers.GetCertificates)
	certificates.Get("/:certificateId", controllers.GetCertificateDetail)
}
