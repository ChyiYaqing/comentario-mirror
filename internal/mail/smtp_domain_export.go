package mail

import (
	"bytes"
	"gitlab.com/comentario/comentario/internal/config"
	"gitlab.com/comentario/comentario/internal/util"
	"net/smtp"
	"os"
)

type domainExportPlugs struct {
	Origin    string
	Domain    string
	ExportHex string
}

func SMTPDomainExport(to string, toName string, _ string, exportHex string) error {
	var header bytes.Buffer
	if err := headerTemplate.Execute(&header, &headerPlugs{FromAddress: config.CLIFlags.EmailFrom, ToAddress: to, ToName: toName, Subject: "Comentario Data Export"}); err != nil {
		return err
	}

	var body bytes.Buffer
	if err := templates["domain-export"].Execute(&body, &domainExportPlugs{Origin: os.Getenv("ORIGIN"), ExportHex: exportHex}); err != nil {
		return err
	}

	err := smtp.SendMail(os.Getenv("SMTP_HOST")+":"+os.Getenv("SMTP_PORT"), smtpAuth, config.CLIFlags.EmailFrom, []string{to}, concat(header, body))
	if err != nil {
		logger.Errorf("cannot send data export email: %v", err)
		return util.ErrorCannotSendEmail
	}

	return nil
}
