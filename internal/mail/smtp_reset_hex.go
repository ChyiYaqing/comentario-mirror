package mail

import (
	"bytes"
	"gitlab.com/commento/commento/api/internal/util"
	"net/smtp"
	"os"
)

type resetHexPlugs struct {
	Origin   string
	ResetHex string
}

func SMTPResetHex(to string, toName string, resetHex string) error {
	var header bytes.Buffer
	if err := headerTemplate.Execute(&header, &headerPlugs{FromAddress: os.Getenv("SMTP_FROM_ADDRESS"), ToAddress: to, ToName: toName, Subject: "Reset your password"}); err != nil {
		return err
	}

	var body bytes.Buffer
	if err := templates["reset-hex"].Execute(&body, &resetHexPlugs{Origin: os.Getenv("ORIGIN"), ResetHex: resetHex}); err != nil {
		return err
	}

	err := smtp.SendMail(os.Getenv("SMTP_HOST")+":"+os.Getenv("SMTP_PORT"), smtpAuth, os.Getenv("SMTP_FROM_ADDRESS"), []string{to}, concat(header, body))
	if err != nil {
		logger.Errorf("cannot send reset email: %v", err)
		return util.ErrorCannotSendEmail
	}

	return nil
}
