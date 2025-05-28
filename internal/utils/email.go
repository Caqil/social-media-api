// utils/email.go
package utils

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
	"os"
	"strings"
	"time"
)

// EmailConfig holds email configuration
type EmailConfig struct {
	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string
	FromEmail    string
	FromName     string
}

// EmailData represents data for email templates
type EmailData struct {
	UserName   string
	AppName    string
	AppURL     string
	Subject    string
	Content    string
	ActionURL  string
	ActionText string
	Year       int
	ExpiryTime string
	Token      string
	Code       string
	// Additional fields for specific email types
	PostTitle     string
	CommentText   string
	GroupName     string
	EventTitle    string
	SenderName    string
	RecipientName string
}

// EmailService handles email operations
type EmailService struct {
	config EmailConfig
}

// NewEmailService creates a new email service instance
func NewEmailService() *EmailService {
	config := EmailConfig{
		SMTPHost:     getEnv("SMTP_HOST", "localhost"),
		SMTPPort:     getEnv("SMTP_PORT", "587"),
		SMTPUser:     getEnv("SMTP_USER", ""),
		SMTPPassword: getEnv("SMTP_PASS", ""),
		FromEmail:    getEnv("FROM_EMAIL", "noreply@socialmedia.com"),
		FromName:     getEnv("FROM_NAME", "Social Media App"),
	}

	return &EmailService{config: config}
}

// SendEmail sends a basic email
func (es *EmailService) SendEmail(to, subject, body string) error {
	if es.config.SMTPUser == "" || es.config.SMTPPassword == "" {
		return fmt.Errorf("SMTP credentials not configured")
	}

	// Set up authentication
	auth := smtp.PlainAuth("", es.config.SMTPUser, es.config.SMTPPassword, es.config.SMTPHost)

	// Compose message
	from := fmt.Sprintf("%s <%s>", es.config.FromName, es.config.FromEmail)
	msg := []byte(fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		from, to, subject, body,
	))

	// Send email
	addr := es.config.SMTPHost + ":" + es.config.SMTPPort
	return smtp.SendMail(addr, auth, es.config.FromEmail, []string{to}, msg)
}

// SendWelcomeEmail sends welcome email to new users
func (es *EmailService) SendWelcomeEmail(to, userName string) error {
	data := EmailData{
		UserName: userName,
		AppName:  AppName,
		AppURL:   getEnv("FRONTEND_URL", "http://localhost:3000"),
		Subject:  "Welcome to " + AppName,
		Year:     time.Now().Year(),
	}

	body, err := es.renderTemplate("welcome", data)
	if err != nil {
		return err
	}

	return es.SendEmail(to, data.Subject, body)
}

// SendEmailVerification sends email verification link
func (es *EmailService) SendEmailVerification(to, userName, token string) error {
	frontendURL := getEnv("FRONTEND_URL", "http://localhost:3000")
	actionURL := fmt.Sprintf("%s/verify-email?token=%s", frontendURL, token)

	data := EmailData{
		UserName:   userName,
		AppName:    AppName,
		AppURL:     frontendURL,
		Subject:    "Verify your email address",
		ActionURL:  actionURL,
		ActionText: "Verify Email",
		Token:      token,
		ExpiryTime: "24 hours",
		Year:       time.Now().Year(),
	}

	body, err := es.renderTemplate("email_verification", data)
	if err != nil {
		return err
	}

	return es.SendEmail(to, data.Subject, body)
}

// SendPasswordReset sends password reset email
func (es *EmailService) SendPasswordReset(to, userName, token string) error {
	frontendURL := getEnv("FRONTEND_URL", "http://localhost:3000")
	actionURL := fmt.Sprintf("%s/reset-password?token=%s", frontendURL, token)

	data := EmailData{
		UserName:   userName,
		AppName:    AppName,
		AppURL:     frontendURL,
		Subject:    "Reset your password",
		ActionURL:  actionURL,
		ActionText: "Reset Password",
		Token:      token,
		ExpiryTime: "1 hour",
		Year:       time.Now().Year(),
	}

	body, err := es.renderTemplate("password_reset", data)
	if err != nil {
		return err
	}

	return es.SendEmail(to, data.Subject, body)
}

// SendNotificationEmail sends notification emails
func (es *EmailService) SendNotificationEmail(to, userName, title, content, actionURL string) error {
	data := EmailData{
		UserName:   userName,
		AppName:    AppName,
		AppURL:     getEnv("FRONTEND_URL", "http://localhost:3000"),
		Subject:    title,
		Content:    content,
		ActionURL:  actionURL,
		ActionText: "View",
		Year:       time.Now().Year(),
	}

	body, err := es.renderTemplate("notification", data)
	if err != nil {
		return err
	}

	return es.SendEmail(to, data.Subject, body)
}

// SendGroupInvitation sends group invitation email
func (es *EmailService) SendGroupInvitation(to, userName, groupName, inviterName, groupURL string) error {
	data := EmailData{
		UserName:   userName,
		AppName:    AppName,
		AppURL:     getEnv("FRONTEND_URL", "http://localhost:3000"),
		Subject:    fmt.Sprintf("You've been invited to join %s", groupName),
		GroupName:  groupName,
		SenderName: inviterName,
		ActionURL:  groupURL,
		ActionText: "View Group",
		Year:       time.Now().Year(),
	}

	body, err := es.renderTemplate("group_invitation", data)
	if err != nil {
		return err
	}

	return es.SendEmail(to, data.Subject, body)
}

// SendEventInvitation sends event invitation email
func (es *EmailService) SendEventInvitation(to, userName, eventTitle, inviterName, eventURL string) error {
	data := EmailData{
		UserName:   userName,
		AppName:    AppName,
		AppURL:     getEnv("FRONTEND_URL", "http://localhost:3000"),
		Subject:    fmt.Sprintf("You're invited to %s", eventTitle),
		EventTitle: eventTitle,
		SenderName: inviterName,
		ActionURL:  eventURL,
		ActionText: "View Event",
		Year:       time.Now().Year(),
	}

	body, err := es.renderTemplate("event_invitation", data)
	if err != nil {
		return err
	}

	return es.SendEmail(to, data.Subject, body)
}

// renderTemplate renders email template with data
func (es *EmailService) renderTemplate(templateName string, data EmailData) (string, error) {
	templates := map[string]string{
		"welcome": `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>{{.Subject}}</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 0; padding: 20px; background-color: #f4f4f4; }
        .container { max-width: 600px; margin: 0 auto; background-color: white; padding: 30px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .header { text-align: center; margin-bottom: 30px; }
        .logo { font-size: 24px; font-weight: bold; color: #333; }
        .content { line-height: 1.6; color: #666; }
        .footer { margin-top: 30px; padding-top: 20px; border-top: 1px solid #eee; text-align: center; color: #999; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1 class="logo">{{.AppName}}</h1>
        </div>
        <div class="content">
            <h2>Welcome, {{.UserName}}!</h2>
            <p>Thank you for joining our community. We're excited to have you on board!</p>
            <p>You can now start sharing posts, connecting with friends, and exploring all the features we have to offer.</p>
            <p>If you have any questions, feel free to reach out to our support team.</p>
        </div>
        <div class="footer">
            <p>&copy; {{.Year}} {{.AppName}}. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`,

		"email_verification": `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>{{.Subject}}</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 0; padding: 20px; background-color: #f4f4f4; }
        .container { max-width: 600px; margin: 0 auto; background-color: white; padding: 30px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .header { text-align: center; margin-bottom: 30px; }
        .logo { font-size: 24px; font-weight: bold; color: #333; }
        .content { line-height: 1.6; color: #666; }
        .button { display: inline-block; padding: 12px 24px; background-color: #007bff; color: white; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .footer { margin-top: 30px; padding-top: 20px; border-top: 1px solid #eee; text-align: center; color: #999; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1 class="logo">{{.AppName}}</h1>
        </div>
        <div class="content">
            <h2>Verify your email address</h2>
            <p>Hi {{.UserName}},</p>
            <p>Please click the button below to verify your email address. This link will expire in {{.ExpiryTime}}.</p>
            <p style="text-align: center;">
                <a href="{{.ActionURL}}" class="button">{{.ActionText}}</a>
            </p>
            <p>If you didn't create an account, you can safely ignore this email.</p>
        </div>
        <div class="footer">
            <p>&copy; {{.Year}} {{.AppName}}. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`,

		"password_reset": `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>{{.Subject}}</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 0; padding: 20px; background-color: #f4f4f4; }
        .container { max-width: 600px; margin: 0 auto; background-color: white; padding: 30px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .header { text-align: center; margin-bottom: 30px; }
        .logo { font-size: 24px; font-weight: bold; color: #333; }
        .content { line-height: 1.6; color: #666; }
        .button { display: inline-block; padding: 12px 24px; background-color: #dc3545; color: white; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .footer { margin-top: 30px; padding-top: 20px; border-top: 1px solid #eee; text-align: center; color: #999; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1 class="logo">{{.AppName}}</h1>
        </div>
        <div class="content">
            <h2>Reset your password</h2>
            <p>Hi {{.UserName}},</p>
            <p>We received a request to reset your password. Click the button below to create a new password. This link will expire in {{.ExpiryTime}}.</p>
            <p style="text-align: center;">
                <a href="{{.ActionURL}}" class="button">{{.ActionText}}</a>
            </p>
            <p>If you didn't request a password reset, you can safely ignore this email.</p>
        </div>
        <div class="footer">
            <p>&copy; {{.Year}} {{.AppName}}. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`,

		"notification": `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>{{.Subject}}</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 0; padding: 20px; background-color: #f4f4f4; }
        .container { max-width: 600px; margin: 0 auto; background-color: white; padding: 30px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .header { text-align: center; margin-bottom: 30px; }
        .logo { font-size: 24px; font-weight: bold; color: #333; }
        .content { line-height: 1.6; color: #666; }
        .button { display: inline-block; padding: 12px 24px; background-color: #28a745; color: white; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .footer { margin-top: 30px; padding-top: 20px; border-top: 1px solid #eee; text-align: center; color: #999; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1 class="logo">{{.AppName}}</h1>
        </div>
        <div class="content">
            <h2>{{.Subject}}</h2>
            <p>Hi {{.UserName}},</p>
            <p>{{.Content}}</p>
            {{if .ActionURL}}
            <p style="text-align: center;">
                <a href="{{.ActionURL}}" class="button">{{.ActionText}}</a>
            </p>
            {{end}}
        </div>
        <div class="footer">
            <p>&copy; {{.Year}} {{.AppName}}. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`,

		"group_invitation": `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>{{.Subject}}</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 0; padding: 20px; background-color: #f4f4f4; }
        .container { max-width: 600px; margin: 0 auto; background-color: white; padding: 30px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .header { text-align: center; margin-bottom: 30px; }
        .logo { font-size: 24px; font-weight: bold; color: #333; }
        .content { line-height: 1.6; color: #666; }
        .button { display: inline-block; padding: 12px 24px; background-color: #6f42c1; color: white; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .footer { margin-top: 30px; padding-top: 20px; border-top: 1px solid #eee; text-align: center; color: #999; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1 class="logo">{{.AppName}}</h1>
        </div>
        <div class="content">
            <h2>{{.Subject}}</h2>
            <p>Hi {{.UserName}},</p>
            <p>{{.SenderName}} has invited you to join the group "{{.GroupName}}".</p>
            <p>Click the button below to view the group and decide if you'd like to join.</p>
            <p style="text-align: center;">
                <a href="{{.ActionURL}}" class="button">{{.ActionText}}</a>
            </p>
        </div>
        <div class="footer">
            <p>&copy; {{.Year}} {{.AppName}}. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`,

		"event_invitation": `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>{{.Subject}}</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 0; padding: 20px; background-color: #f4f4f4; }
        .container { max-width: 600px; margin: 0 auto; background-color: white; padding: 30px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .header { text-align: center; margin-bottom: 30px; }
        .logo { font-size: 24px; font-weight: bold; color: #333; }
        .content { line-height: 1.6; color: #666; }
        .button { display: inline-block; padding: 12px 24px; background-color: #fd7e14; color: white; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .footer { margin-top: 30px; padding-top: 20px; border-top: 1px solid #eee; text-align: center; color: #999; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1 class="logo">{{.AppName}}</h1>
        </div>
        <div class="content">
            <h2>{{.Subject}}</h2>
            <p>Hi {{.UserName}},</p>
            <p>{{.SenderName}} has invited you to the event "{{.EventTitle}}".</p>
            <p>Click the button below to view the event details and RSVP.</p>
            <p style="text-align: center;">
                <a href="{{.ActionURL}}" class="button">{{.ActionText}}</a>
            </p>
        </div>
        <div class="footer">
            <p>&copy; {{.Year}} {{.AppName}}. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`,
	}

	tmplContent, exists := templates[templateName]
	if !exists {
		return "", fmt.Errorf("template %s not found", templateName)
	}

	tmpl, err := template.New(templateName).Parse(tmplContent)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// ValidateEmail validates email format
func ValidateEmail(email string) bool {
	if len(email) < 3 || len(email) > 254 {
		return false
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}

	// Basic validation
	local, domain := parts[0], parts[1]

	if len(local) < 1 || len(local) > 64 {
		return false
	}

	if len(domain) < 1 || len(domain) > 253 {
		return false
	}

	if strings.Contains(local, "..") || strings.Contains(domain, "..") {
		return false
	}

	if strings.HasPrefix(local, ".") || strings.HasSuffix(local, ".") {
		return false
	}

	return strings.Contains(domain, ".")
}

// SanitizeEmail sanitizes and normalizes email address
func SanitizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// getEnv gets environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
