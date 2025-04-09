// utils/utils.go
package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/smtp"
	"os"
)

// GenerateRandomString creates a URL-safe random string of specified length
func GenerateRandomString(length int) string {
	byteLength := (length * 6) / 8 // Calculate required bytes for desired length
	b := make([]byte, byteLength)

	_, err := rand.Read(b)
	if err != nil {
		log.Printf("Error generating random bytes: %v", err)
		return ""
	}

	encoding := base64.URLEncoding.WithPadding(base64.NoPadding)
	return encoding.EncodeToString(b)[:length]
}

// SendEmail sends an email using SMTP configuration from environment variables
func SendEmail(to, subject, body string) error {
	// Get SMTP configuration from environment
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USERNAME")
	smtpPass := os.Getenv("SMTP_PASSWORD")
	from := os.Getenv("EMAIL_FROM")

	// Set up authentication
	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)

	// Construct RFC-compliant email message
	msg := []byte(
		"From: " + from + "\r\n" +
			"To: " + to + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: text/html; charset=UTF-8\r\n\r\n" +
			body + "\r\n")

	// Send email
	err := smtp.SendMail(
		smtpHost+":"+smtpPort,
		auth,
		from,
		[]string{to},
		msg,
	)

	if err != nil {
		log.Printf("Failed to send email to %s: %v", to, err)
		return err
	}

	return nil
}

// SendVerificationEmail sends email verification link
func SendVerificationEmail(email, token string) {
	appURL := os.Getenv("APP_URL")
	verificationLink := fmt.Sprintf("%s/verify-email?token=%s", appURL, token)

	htmlBody := fmt.Sprintf(`
		<html>
			<body>
				<h1>Email Verification</h1>
				<p>Please click the link below to verify your email address:</p>
				<p><a href="%s">%s</a></p>
				<p>If you didn't request this, you can safely ignore this email.</p>
			</body>
		</html>
	`, verificationLink, verificationLink)

	if err := SendEmail(email, "Verify Your Email Address", htmlBody); err != nil {
		log.Printf("Failed to send verification email: %v", err)
	}
}

// SendPasswordResetEmail sends password reset instructions
func SendPasswordResetEmail(email, token string) {
	appURL := os.Getenv("APP_URL")
	resetLink := fmt.Sprintf("%s/reset-password?token=%s", appURL, token)

	htmlBody := fmt.Sprintf(`
		<html>
			<body>
				<h1>Password Reset Request</h1>
				<p>Click the link below to reset your password. This link will expire in 1 hour.</p>
				<p><a href="%s">%s</a></p>
				<p>If you didn't request this, please secure your account.</p>
			</body>
		</html>
	`, resetLink, resetLink)

	if err := SendEmail(email, "Password Reset Instructions", htmlBody); err != nil {
		log.Printf("Failed to send password reset email: %v", err)
	}
}
