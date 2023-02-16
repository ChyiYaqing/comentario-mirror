package mail

import (
	"bytes"
	"gitlab.com/comentario/comentario/internal/config"
	"gitlab.com/comentario/comentario/internal/util"
	"net/smtp"
	"os"
)

func SMTPDomainExportError(to string, toName string, _ string) error {
	var header bytes.Buffer
	if err := headerTemplate.Execute(&header, &headerPlugs{FromAddress: config.CLIFlags.EmailFrom, ToAddress: to, ToName: toName, Subject: "Comentario Data Export"}); err != nil {
		return err
	}

	var body bytes.Buffer
	if err := templates["data-export-error"].Execute(&body, &domainExportPlugs{Origin: os.Getenv("ORIGIN")}); err != nil {
		return err
	}

	err := smtp.SendMail(os.Getenv("SMTP_HOST")+":"+os.Getenv("SMTP_PORT"), smtpAuth, config.CLIFlags.EmailFrom, []string{to}, concat(header, body))
	if err != nil {
		logger.Errorf("cannot send data export error email: %v", err)
		return util.ErrorCannotSendEmail
	}

	return nil
}
