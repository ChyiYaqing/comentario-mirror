package util

import (
	"gopkg.in/gomail.v2"
)

// NewSMTPMailer instantiates a new Mailer capable of sending out emails using SMTP
func NewSMTPMailer(host string, port int, username, password, emailFrom string) Mailer {
	return &smtpMailer{
		emailFrom: emailFrom,
		dialer:    gomail.NewDialer(host, port, username, password),
	}
}

// smtpMailer is a Mailer implementation that sends emails using the specified SMTP server
type smtpMailer struct {
	emailFrom string
	dialer    *gomail.Dialer
}

func (m *smtpMailer) Mail(replyTo, recipient, subject, htmlMessage string) error {
	// Compose an email
	msg := gomail.NewMessage()
	msg.SetHeader("From", m.emailFrom)
	msg.SetHeader("To", recipient)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/html", htmlMessage)
	if replyTo != "" {
		msg.SetHeader("Reply-To", replyTo)
	}

	// Send it out
	return m.dialer.DialAndSend(msg)
}
