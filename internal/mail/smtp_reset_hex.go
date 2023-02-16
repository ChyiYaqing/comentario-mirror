package mail

import (
	"bytes"
	"gitlab.com/comentario/comentario/internal/config"
	"gitlab.com/comentario/comentario/internal/util"
	"net/smtp"
	"os"
)

type resetHexPlugs struct {
	Origin   string
	ResetHex string
}

func SMTPResetHex(to string, toName string, resetHex string) error {
	var header bytes.Buffer
	if err := headerTemplate.Execute(&header, &headerPlugs{FromAddress: config.CLIFlags.EmailFrom, ToAddress: to, ToName: toName, Subject: "Reset your password"}); err != nil {
		return err
	}

	var body bytes.Buffer
	if err := templates["reset-hex"].Execute(&body, &resetHexPlugs{Origin: os.Getenv("ORIGIN"), ResetHex: resetHex}); err != nil {
		return err
	}

	err := smtp.SendMail(os.Getenv("SMTP_HOST")+":"+os.Getenv("SMTP_PORT"), smtpAuth, config.CLIFlags.EmailFrom, []string{to}, concat(header, body))
	if err != nil {
		logger.Errorf("cannot send reset email: %v", err)
		return util.ErrorCannotSendEmail
	}

	return nil
}
