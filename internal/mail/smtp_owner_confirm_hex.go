package mail

import (
	"bytes"
	"gitlab.com/comentario/comentario/internal/util"
	"net/smtp"
	"os"
)

type ownerConfirmHexPlugs struct {
	Origin     string
	ConfirmHex string
}

func SMTPOwnerConfirmHex(to string, toName string, confirmHex string) error {
	var header bytes.Buffer
	if err := headerTemplate.Execute(&header, &headerPlugs{FromAddress: os.Getenv("SMTP_FROM_ADDRESS"), ToAddress: to, ToName: toName, Subject: "Please confirm your email address"}); err != nil {
		return err
	}

	var body bytes.Buffer
	if err := templates["confirm-hex"].Execute(&body, &ownerConfirmHexPlugs{Origin: os.Getenv("ORIGIN"), ConfirmHex: confirmHex}); err != nil {
		return err
	}

	err := smtp.SendMail(os.Getenv("SMTP_HOST")+":"+os.Getenv("SMTP_PORT"), smtpAuth, os.Getenv("SMTP_FROM_ADDRESS"), []string{to}, concat(header, body))
	if err != nil {
		logger.Errorf("cannot send confirmation email: %v", err)
		return util.ErrorCannotSendEmail
	}

	return nil
}
