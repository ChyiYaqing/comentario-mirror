package main

import (
	"bytes"
	"net/smtp"
	"os"
)

type domainExportErrorPlugs struct {
	Origin string
	Domain string
}

func smtpDomainExportError(to string, toName string, _ string) error {
	var header bytes.Buffer
	if err := headerTemplate.Execute(&header, &headerPlugs{FromAddress: os.Getenv("SMTP_FROM_ADDRESS"), ToAddress: to, ToName: toName, Subject: "Commento Data Export"}); err != nil {
		return err
	}

	var body bytes.Buffer
	if err := templates["data-export-error"].Execute(&body, &domainExportPlugs{Origin: os.Getenv("ORIGIN")}); err != nil {
		return err
	}

	err := smtp.SendMail(os.Getenv("SMTP_HOST")+":"+os.Getenv("SMTP_PORT"), smtpAuth, os.Getenv("SMTP_FROM_ADDRESS"), []string{to}, concat(header, body))
	if err != nil {
		logger.Errorf("cannot send data export error email: %v", err)
		return errorCannotSendEmail
	}

	return nil
}
