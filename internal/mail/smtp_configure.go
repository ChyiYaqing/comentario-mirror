package mail

import (
	"bytes"
	"github.com/op/go-logging"
	"gitlab.com/comentario/comentario/internal/util"
	"net/smtp"
	"os"
)

var logger = logging.MustGetLogger("mail")
var SMTPConfigured bool
var smtpAuth smtp.Auth

func SMTPConfigure() error {
	username := os.Getenv("SMTP_USERNAME")
	password := os.Getenv("SMTP_PASSWORD")
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	if host == "" || port == "" {
		logger.Warningf("smtp not configured, no emails will be sent")
		SMTPConfigured = false
		return nil
	}

	if os.Getenv("SMTP_FROM_ADDRESS") == "" {
		logger.Errorf("COMENTARIO_SMTP_FROM_ADDRESS not set")
		SMTPConfigured = false
		return util.ErrorMissingSmtpAddress
	}

	logger.Infof("configuring smtp: %s", host)
	if username == "" || password == "" {
		logger.Warningf("no SMTP username/password set, Comentario will assume they aren't required")
	} else {
		smtpAuth = smtp.PlainAuth("", username, password, host)
	}
	SMTPConfigured = true
	return nil
}

func concat(a bytes.Buffer, b bytes.Buffer) []byte {
	return append(a.Bytes(), b.Bytes()...)
}
