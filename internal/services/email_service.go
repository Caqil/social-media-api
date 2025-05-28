// internal/services/email_service.go
package services

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/smtp"
	"strings"
	"time"

	"social-media-api/internal/models"
)

type EmailService struct {
	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
	FromName     string
	Templates    map[string]*template.Template
}

type EmailData struct {
	To          []string
	Subject     string
	Body        string
	HTMLBody    string
	Attachments []EmailAttachment
}

type EmailAttachment struct {
	Filename string
	Content  []byte
	MimeType string
}

func NewEmailService(smtpHost, smtpPort, smtpUsername, smtpPassword, fromEmail, fromName string) *EmailService {
	es := &EmailService{
		SMTPHost:     smtpHost,
		SMTPPort:     smtpPort,
		SMTPUsername: smtpUsername,
		SMTPPassword: smtpPassword,
		FromEmail:    fromEmail,
		FromName:     fromName,
		Templates:    make(map[string]*template.Template),
	}

	// Load email templates
	es.loadTemplates()
	return es
}

// SendEmail sends a basic email
func (es *EmailService) SendEmail(data EmailData) error {
	auth := smtp.PlainAuth("", es.SMTPUsername, es.SMTPPassword, es.SMTPHost)

	// Create message
	msg := es.buildMessage(data)

	// Send email
	err := smtp.SendMail(
		es.SMTPHost+":"+es.SMTPPort,
		auth,
		es.FromEmail,
		data.To,
		[]byte(msg),
	)

	if err != nil {
		log.Printf("Failed to send email: %v", err)
		return err
	}

	log.Printf("Email sent successfully to: %v", data.To)
	return nil
}

// SendWelcomeEmail sends welcome email to new users
func (es *EmailService) SendWelcomeEmail(user *models.User, verificationToken string) error {
	data := map[string]interface{}{
		"User":              user,
		"VerificationToken": verificationToken,
		"AppName":           "Social Media App",
		"SupportEmail":      "support@example.com",
		"Year":              time.Now().Year(),
	}

	htmlBody, err := es.renderTemplate("welcome", data)
	if err != nil {
		return err
	}

	emailData := EmailData{
		To:       []string{user.Email},
		Subject:  "Welcome to Social Media App!",
		HTMLBody: htmlBody,
		Body:     es.generatePlainTextVersion(htmlBody),
	}

	return es.SendEmail(emailData)
}

// SendEmailVerification sends email verification
func (es *EmailService) SendEmailVerification(user *models.User, verificationToken string) error {
	data := map[string]interface{}{
		"User":              user,
		"VerificationToken": verificationToken,
		"AppName":           "Social Media App",
		"Year":              time.Now().Year(),
	}

	htmlBody, err := es.renderTemplate("email_verification", data)
	if err != nil {
		return err
	}

	emailData := EmailData{
		To:       []string{user.Email},
		Subject:  "Verify Your Email Address",
		HTMLBody: htmlBody,
		Body:     es.generatePlainTextVersion(htmlBody),
	}

	return es.SendEmail(emailData)
}

// SendPasswordResetEmail sends password reset email
func (es *EmailService) SendPasswordResetEmail(user *models.User, resetToken string) error {
	data := map[string]interface{}{
		"User":       user,
		"ResetToken": resetToken,
		"AppName":    "Social Media App",
		"Year":       time.Now().Year(),
	}

	htmlBody, err := es.renderTemplate("password_reset", data)
	if err != nil {
		return err
	}

	emailData := EmailData{
		To:       []string{user.Email},
		Subject:  "Reset Your Password",
		HTMLBody: htmlBody,
		Body:     es.generatePlainTextVersion(htmlBody),
	}

	return es.SendEmail(emailData)
}

// SendPasswordChangeConfirmation sends password change confirmation
func (es *EmailService) SendPasswordChangeConfirmation(user *models.User) error {
	data := map[string]interface{}{
		"User":    user,
		"AppName": "Social Media App",
		"Year":    time.Now().Year(),
	}

	htmlBody, err := es.renderTemplate("password_changed", data)
	if err != nil {
		return err
	}

	emailData := EmailData{
		To:       []string{user.Email},
		Subject:  "Your Password Has Been Changed",
		HTMLBody: htmlBody,
		Body:     es.generatePlainTextVersion(htmlBody),
	}

	return es.SendEmail(emailData)
}

// SendNotificationEmail sends notification emails
func (es *EmailService) SendNotificationEmail(notification *models.Notification) error {
	// Get recipient user info (this would be passed or fetched)
	recipientEmail := "user@example.com" // This should be fetched from user service

	data := map[string]interface{}{
		"Notification": notification,
		"AppName":      "Social Media App",
		"Year":         time.Now().Year(),
	}

	htmlBody, err := es.renderTemplate("notification", data)
	if err != nil {
		return err
	}

	emailData := EmailData{
		To:       []string{recipientEmail},
		Subject:  notification.Title,
		HTMLBody: htmlBody,
		Body:     es.generatePlainTextVersion(htmlBody),
	}

	return es.SendEmail(emailData)
}

// SendDigestEmail sends daily/weekly digest emails
func (es *EmailService) SendDigestEmail(user *models.User, digestData interface{}, digestType string) error {
	var subject string
	switch digestType {
	case "daily":
		subject = "Your Daily Digest"
	case "weekly":
		subject = "Your Weekly Digest"
	default:
		subject = "Your Digest"
	}

	data := map[string]interface{}{
		"User":       user,
		"DigestData": digestData,
		"DigestType": digestType,
		"AppName":    "Social Media App",
		"Year":       time.Now().Year(),
	}

	htmlBody, err := es.renderTemplate("digest", data)
	if err != nil {
		return err
	}

	emailData := EmailData{
		To:       []string{user.Email},
		Subject:  subject,
		HTMLBody: htmlBody,
		Body:     es.generatePlainTextVersion(htmlBody),
	}

	return es.SendEmail(emailData)
}

// SendAccountSuspensionEmail sends account suspension notification
func (es *EmailService) SendAccountSuspensionEmail(user *models.User, reason string) error {
	data := map[string]interface{}{
		"User":         user,
		"Reason":       reason,
		"AppName":      "Social Media App",
		"SupportEmail": "support@example.com",
		"Year":         time.Now().Year(),
	}

	htmlBody, err := es.renderTemplate("account_suspended", data)
	if err != nil {
		return err
	}

	emailData := EmailData{
		To:       []string{user.Email},
		Subject:  "Your Account Has Been Suspended",
		HTMLBody: htmlBody,
		Body:     es.generatePlainTextVersion(htmlBody),
	}

	return es.SendEmail(emailData)
}

// SendGroupInviteEmail sends group invitation email
func (es *EmailService) SendGroupInviteEmail(invitee *models.User, group interface{}, inviter *models.User) error {
	data := map[string]interface{}{
		"Invitee": invitee,
		"Group":   group,
		"Inviter": inviter,
		"AppName": "Social Media App",
		"Year":    time.Now().Year(),
	}

	htmlBody, err := es.renderTemplate("group_invite", data)
	if err != nil {
		return err
	}

	emailData := EmailData{
		To:       []string{invitee.Email},
		Subject:  "You've been invited to join a group",
		HTMLBody: htmlBody,
		Body:     es.generatePlainTextVersion(htmlBody),
	}

	return es.SendEmail(emailData)
}

// SendEventInviteEmail sends event invitation email
func (es *EmailService) SendEventInviteEmail(invitee *models.User, event interface{}, inviter *models.User) error {
	data := map[string]interface{}{
		"Invitee": invitee,
		"Event":   event,
		"Inviter": inviter,
		"AppName": "Social Media App",
		"Year":    time.Now().Year(),
	}

	htmlBody, err := es.renderTemplate("event_invite", data)
	if err != nil {
		return err
	}

	emailData := EmailData{
		To:       []string{invitee.Email},
		Subject:  "You've been invited to an event",
		HTMLBody: htmlBody,
		Body:     es.generatePlainTextVersion(htmlBody),
	}

	return es.SendEmail(emailData)
}

// SendEventReminderEmail sends event reminder email
func (es *EmailService) SendEventReminderEmail(user *models.User, event interface{}) error {
	data := map[string]interface{}{
		"User":    user,
		"Event":   event,
		"AppName": "Social Media App",
		"Year":    time.Now().Year(),
	}

	htmlBody, err := es.renderTemplate("event_reminder", data)
	if err != nil {
		return err
	}

	emailData := EmailData{
		To:       []string{user.Email},
		Subject:  "Event Reminder",
		HTMLBody: htmlBody,
		Body:     es.generatePlainTextVersion(htmlBody),
	}

	return es.SendEmail(emailData)
}

// SendSecurityAlertEmail sends security alert emails
func (es *EmailService) SendSecurityAlertEmail(user *models.User, alertType, details string) error {
	data := map[string]interface{}{
		"User":         user,
		"AlertType":    alertType,
		"Details":      details,
		"AppName":      "Social Media App",
		"SupportEmail": "support@example.com",
		"Year":         time.Now().Year(),
	}

	htmlBody, err := es.renderTemplate("security_alert", data)
	if err != nil {
		return err
	}

	emailData := EmailData{
		To:       []string{user.Email},
		Subject:  "Security Alert - " + alertType,
		HTMLBody: htmlBody,
		Body:     es.generatePlainTextVersion(htmlBody),
	}

	return es.SendEmail(emailData)
}

// Helper methods

func (es *EmailService) buildMessage(data EmailData) string {
	var msg bytes.Buffer

	// Headers
	msg.WriteString(fmt.Sprintf("From: %s <%s>\r\n", es.FromName, es.FromEmail))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(data.To, ",")))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", data.Subject))
	msg.WriteString("MIME-Version: 1.0\r\n")

	if data.HTMLBody != "" {
		// Multipart message
		boundary := "boundary-" + fmt.Sprintf("%d", time.Now().Unix())
		msg.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=%s\r\n\r\n", boundary))

		// Plain text part
		msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		msg.WriteString("Content-Type: text/plain; charset=UTF-8\r\n\r\n")
		if data.Body != "" {
			msg.WriteString(data.Body)
		} else {
			msg.WriteString(es.generatePlainTextVersion(data.HTMLBody))
		}
		msg.WriteString("\r\n\r\n")

		// HTML part
		msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		msg.WriteString("Content-Type: text/html; charset=UTF-8\r\n\r\n")
		msg.WriteString(data.HTMLBody)
		msg.WriteString("\r\n\r\n")

		msg.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
	} else {
		// Plain text only
		msg.WriteString("Content-Type: text/plain; charset=UTF-8\r\n\r\n")
		msg.WriteString(data.Body)
	}

	return msg.String()
}

func (es *EmailService) renderTemplate(templateName string, data interface{}) (string, error) {
	tmpl, exists := es.Templates[templateName]
	if !exists {
		return "", fmt.Errorf("template %s not found", templateName)
	}

	var buf bytes.Buffer
	err := tmpl.Execute(&buf, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (es *EmailService) generatePlainTextVersion(htmlBody string) string {
	// Simple HTML to plain text conversion
	// In production, you might want to use a proper HTML-to-text library
	plainText := strings.ReplaceAll(htmlBody, "<br>", "\n")
	plainText = strings.ReplaceAll(plainText, "<br/>", "\n")
	plainText = strings.ReplaceAll(plainText, "<br />", "\n")
	plainText = strings.ReplaceAll(plainText, "</p>", "\n\n")
	plainText = strings.ReplaceAll(plainText, "</div>", "\n")

	// Remove all HTML tags
	for strings.Contains(plainText, "<") && strings.Contains(plainText, ">") {
		start := strings.Index(plainText, "<")
		end := strings.Index(plainText[start:], ">") + start + 1
		if end > start {
			plainText = plainText[:start] + plainText[end:]
		} else {
			break
		}
	}

	// Clean up extra whitespace
	lines := strings.Split(plainText, "\n")
	var cleanLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			cleanLines = append(cleanLines, line)
		}
	}

	return strings.Join(cleanLines, "\n")
}

func (es *EmailService) loadTemplates() {
	// Load email templates
	// In production, these would be loaded from files or embedded resources

	// Welcome email template
	welcomeTemplate := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Welcome to {{.AppName}}</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1 style="color: #4CAF50;">Welcome to {{.AppName}}!</h1>
        <p>Hi {{.User.FirstName}},</p>
        <p>Welcome to {{.AppName}}! We're excited to have you join our community.</p>
        <p>To get started, please verify your email address by clicking the button below:</p>
        <div style="text-align: center; margin: 30px 0;">
            <a href="https://yourapp.com/verify-email?token={{.VerificationToken}}" 
               style="background-color: #4CAF50; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; display: inline-block;">
                Verify Email
            </a>
        </div>
        <p>If the button doesn't work, you can copy and paste this link into your browser:</p>
        <p style="word-break: break-all;">https://yourapp.com/verify-email?token={{.VerificationToken}}</p>
        <p>If you have any questions, feel free to contact us at {{.SupportEmail}}.</p>
        <p>Best regards,<br>The {{.AppName}} Team</p>
        <hr>
        <p style="font-size: 12px; color: #666;">© {{.Year}} {{.AppName}}. All rights reserved.</p>
    </div>
</body>
</html>`

	// Email verification template
	emailVerificationTemplate := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Verify Your Email</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1 style="color: #2196F3;">Verify Your Email Address</h1>
        <p>Hi {{.User.FirstName}},</p>
        <p>Please verify your email address by clicking the button below:</p>
        <div style="text-align: center; margin: 30px 0;">
            <a href="https://yourapp.com/verify-email?token={{.VerificationToken}}" 
               style="background-color: #2196F3; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; display: inline-block;">
                Verify Email
            </a>
        </div>
        <p>This link will expire in 24 hours for security reasons.</p>
        <p>If you didn't request this verification, please ignore this email.</p>
        <p>Best regards,<br>The {{.AppName}} Team</p>
        <hr>
        <p style="font-size: 12px; color: #666;">© {{.Year}} {{.AppName}}. All rights reserved.</p>
    </div>
</body>
</html>`

	// Password reset template
	passwordResetTemplate := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Reset Your Password</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1 style="color: #FF9800;">Reset Your Password</h1>
        <p>Hi {{.User.FirstName}},</p>
        <p>We received a request to reset your password. Click the button below to create a new password:</p>
        <div style="text-align: center; margin: 30px 0;">
            <a href="https://yourapp.com/reset-password?token={{.ResetToken}}" 
               style="background-color: #FF9800; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; display: inline-block;">
                Reset Password
            </a>
        </div>
        <p>This link will expire in 1 hour for security reasons.</p>
        <p>If you didn't request a password reset, please ignore this email or contact us if you have concerns.</p>
        <p>Best regards,<br>The {{.AppName}} Team</p>
        <hr>
        <p style="font-size: 12px; color: #666;">© {{.Year}} {{.AppName}}. All rights reserved.</p>
    </div>
</body>
</html>`

	// Notification template
	notificationTemplate := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>{{.Notification.Title}}</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1 style="color: #9C27B0;">{{.Notification.Title}}</h1>
        <p>{{.Notification.Message}}</p>
        {{if .Notification.ActionText}}
        <div style="text-align: center; margin: 30px 0;">
            <a href="https://yourapp.com{{.Notification.TargetURL}}" 
               style="background-color: #9C27B0; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; display: inline-block;">
                {{.Notification.ActionText}}
            </a>
        </div>
        {{end}}
        <p>Best regards,<br>The {{.AppName}} Team</p>
        <hr>
        <p style="font-size: 12px; color: #666;">© {{.Year}} {{.AppName}}. All rights reserved.</p>
    </div>
</body>
</html>`

	// Parse and store templates
	es.Templates["welcome"] = template.Must(template.New("welcome").Parse(welcomeTemplate))
	es.Templates["email_verification"] = template.Must(template.New("email_verification").Parse(emailVerificationTemplate))
	es.Templates["password_reset"] = template.Must(template.New("password_reset").Parse(passwordResetTemplate))
	es.Templates["notification"] = template.Must(template.New("notification").Parse(notificationTemplate))

	// Add more templates as needed
	es.Templates["password_changed"] = es.Templates["notification"]  // Reuse notification template
	es.Templates["account_suspended"] = es.Templates["notification"] // Reuse notification template
	es.Templates["group_invite"] = es.Templates["notification"]      // Reuse notification template
	es.Templates["event_invite"] = es.Templates["notification"]      // Reuse notification template
	es.Templates["event_reminder"] = es.Templates["notification"]    // Reuse notification template
	es.Templates["security_alert"] = es.Templates["notification"]    // Reuse notification template
	es.Templates["digest"] = es.Templates["notification"]            // Reuse notification template
}
