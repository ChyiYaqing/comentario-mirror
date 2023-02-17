package svc

import (
	"bytes"
	"gitlab.com/comentario/comentario/internal/config"
	"gitlab.com/comentario/comentario/internal/util"
	"html/template"
	"path"
)

// TheEmailService is a global EmailService implementation
var TheEmailService EmailService = &emailService{}

// EmailService is a service interface for emailing
type EmailService interface {
	// Send sends an email and logs the outcome
	Send(replyTo, recipient, subject, htmlMessage string) error
	// SendFromTemplate sends an email from the provided template and logs the outcome
	SendFromTemplate(replyTo, recipient, subject, templateFile string, templateData map[string]any) error
}

//----------------------------------------------------------------------------------------------------------------------

// emailService is a blueprint EmailService implementation
type emailService struct{}

func (svc *emailService) Send(replyTo, recipient, subject, htmlMessage string) error {
	err := util.AppMailer.Mail(replyTo, recipient, subject, htmlMessage)
	if err != nil {
		logger.Warningf("Failed to send email to %s: %v", recipient, err)
	} else {
		logger.Debugf("Successfully sent an email to %s", recipient)
	}
	return err
}

func (svc *emailService) SendFromTemplate(replyTo, recipient, subject, templateFile string, templateData map[string]any) error {
	// Load and parse the template
	filename := templatePath(templateFile)
	t, err := template.ParseFiles(filename)
	if err != nil {
		logger.Errorf("Failed to parse HTML template file %s: %v", filename, err)
		return util.ErrorMalformedTemplate
	}
	logger.Debugf("Parsed HTML template %s", filename)

	// Execute the template
	var bufHTML bytes.Buffer
	if err := t.Execute(&bufHTML, templateData); err != nil {
		return err
	}

	// Send the mail
	return svc.Send(replyTo, recipient, subject, bufHTML.String())
}

func templatePath(filename string) string {
	return path.Join(config.CLIFlags.StaticPath, "templates", filename)
}
