package main

import (
	"fmt"
	"net/smtp"
	"strings"
)

const (
	smtpServer = "smtp.gmail.com"
	smtpPort   = "587"
)

type Sender struct {
	gmailAppPassword string
}

func NewSender(gmailAppPassword string) *Sender {
	return &Sender{
		gmailAppPassword: gmailAppPassword,
	}
}

func (s *Sender) SendEmail(subject, textBody, htmlBody string) error {
	auth := smtp.PlainAuth("", EmailAddress, s.gmailAppPassword, smtpServer)
	msg := s.composeMIMEMessage(subject, textBody, htmlBody)
	addr := smtpServer + ":" + smtpPort
	err := smtp.SendMail(addr, auth, EmailAddress, []string{EmailAddress}, []byte(msg))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}

func (s *Sender) composeMIMEMessage(subject, textBody, htmlBody string) string {
	boundary := "----=_Part_0_1234567890.1234567890"

	headers := []string{
		fmt.Sprintf("From: %s", EmailAddress),
		fmt.Sprintf("To: %s", EmailAddress),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
		fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"", boundary),
		"",
	}

	parts := []string{
		strings.Join(headers, "\r\n"),
		fmt.Sprintf("--%s", boundary),
		"Content-Type: text/plain; charset=UTF-8",
		"Content-Transfer-Encoding: 7bit",
		"",
		textBody,
		"",
		fmt.Sprintf("--%s", boundary),
		"Content-Type: text/html; charset=UTF-8",
		"Content-Transfer-Encoding: 7bit",
		"",
		htmlBody,
		"",
		fmt.Sprintf("--%s--", boundary),
	}

	return strings.Join(parts, "\r\n")
}

func FormatRSSEmail(feedName string, item FeedItem) (subject, textBody, htmlBody string) {
	subject = fmt.Sprintf("[RSS] %s: %s", feedName, item.Title)

	textBody = fmt.Sprintf(`
New post from %s

Title: %s
Link: %s
Published: %s

%s

---
This email was sent by RSS to Email service.
`, feedName, item.Title, item.Link, item.Published, item.Summary)

	htmlBody = fmt.Sprintf(`
<html>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <h2 style="color: #2c3e50;">New post from %s</h2>
    
    <div style="background-color: #f8f9fa; padding: 15px; border-left: 4px solid #3498db; margin: 20px 0;">
        <h3 style="margin-top: 0;">
            <a href="%s" style="color: #2980b9; text-decoration: none;">
                %s
            </a>
        </h3>
        <p style="color: #7f8c8d; font-size: 0.9em;">
            <strong>Published:</strong> %s
        </p>
    </div>
    
    <div style="margin: 20px 0;">
        %s
    </div>
    
    <div style="margin-top: 30px; padding-top: 20px; border-top: 1px solid #ecf0f1;">
        <p style="color: #95a5a6; font-size: 0.85em;">
            This email was sent by RSS to Email service.
        </p>
    </div>
</body>
</html>
`, feedName, item.Link, item.Title, item.Published, item.Summary)

	return subject, textBody, htmlBody
}
