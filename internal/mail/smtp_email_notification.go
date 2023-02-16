package mail

import (
	"bytes"
	"fmt"
	"gitlab.com/comentario/comentario/internal/api/models"
	"gitlab.com/comentario/comentario/internal/config"
	"gitlab.com/comentario/comentario/internal/util"
	ht "html/template"
	"net/smtp"
	"os"
	tt "text/template"
)

type emailNotificationPlugs struct {
	Origin               string
	Kind                 string
	UnsubscribeSecretHex models.HexID
	Domain               string
	Path                 string
	CommentHex           models.HexID
	CommenterName        string
	Title                string
	Html                 ht.HTML
}

func SMTPEmailNotification(toEmail string, toName string, kind string, domain string, path string, commentHex models.HexID, commenterName string, title string, html string, unsubscribeSecretHex models.HexID) error {
	h, err := tt.New("header").Parse(`MIME-Version: 1.0
From: Comentario <{{.FromAddress}}>
To: {{.ToName}} <{{.ToAddress}}>
Content-Type: text/html; charset=UTF-8
Subject: {{.Subject}}

`)
	var header bytes.Buffer
	if err := h.Execute(
		&header,
		&headerPlugs{FromAddress: config.CLIFlags.EmailFrom, ToAddress: toEmail, ToName: toName, Subject: "[Comentario] " + title},
	); err != nil {
		return err
	}

	t, err := ht.ParseFiles(fmt.Sprintf("%s/templates/email-notification.txt", os.Getenv("STATIC")))
	if err != nil {
		logger.Errorf("cannot parse %s/templates/email-notification.txt: %v", os.Getenv("STATIC"), err)
		return util.ErrorMalformedTemplate
	}

	var body bytes.Buffer
	err = t.Execute(&body, &emailNotificationPlugs{
		Origin:               os.Getenv("ORIGIN"),
		Kind:                 kind,
		Domain:               domain,
		Path:                 path,
		CommentHex:           commentHex,
		CommenterName:        commenterName,
		Title:                title,
		Html:                 ht.HTML(html),
		UnsubscribeSecretHex: unsubscribeSecretHex,
	})
	if err != nil {
		logger.Errorf("error generating templated HTML for email notification: %v", err)
		return err
	}

	err = smtp.SendMail(os.Getenv("SMTP_HOST")+":"+os.Getenv("SMTP_PORT"), smtpAuth, config.CLIFlags.EmailFrom, []string{toEmail}, concat(header, body))
	if err != nil {
		logger.Errorf("cannot send email notification: %v", err)
		return util.ErrorCannotSendEmail
	}

	return nil
}
