package handlers

import (
	"fmt"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"gitlab.com/comentario/comentario/internal/api/models"
	"gitlab.com/comentario/comentario/internal/api/restapi/operations"
	"gitlab.com/comentario/comentario/internal/mail"
	"gitlab.com/comentario/comentario/internal/svc"
	"gitlab.com/comentario/comentario/internal/util"
	"time"
)

const emailsRowColumns = `
	emails.email,
	emails.unsubscribeSecretHex,
	emails.lastEmailNotificationDate,
	emails.sendReplyNotifications,
	emails.sendModeratorNotifications
`

func EmailGet(params operations.EmailGetParams) middleware.Responder {
	email, err := emailGetByUnsubscribeSecretHex(*params.Body.UnsubscribeSecretHex)
	if err != nil {
		return operations.NewEmailGetOK().WithPayload(&operations.EmailGetOKBody{Message: err.Error()})
	}

	// Succeeded
	return operations.NewEmailGetOK().WithPayload(&operations.EmailGetOKBody{
		Email:   email,
		Success: true,
	})
}

func EmailNew(email strfmt.Email) error {
	unsubscribeSecretHex, err := util.RandomHex(32)
	if err != nil {
		return util.ErrorInternal
	}

	_, err = svc.DB.Exec(
		`insert into emails(email, unsubscribeSecretHex, lastEmailNotificationDate) values ($1, $2, $3) on conflict do nothing;`,
		email,
		unsubscribeSecretHex,
		time.Now().UTC())
	if err != nil {
		logger.Errorf("cannot insert email into emails: %v", err)
		return util.ErrorInternal
	}

	return nil
}

func emailGet(em strfmt.Email) (*models.Email, error) {
	row := svc.DB.QueryRow(
		fmt.Sprintf("select %s from emails where email = $1;", emailsRowColumns),
		em)

	var e models.Email
	if err := emailsRowScan(row, &e); err != nil {
		// TODO: is this the only error?
		return nil, util.ErrorNoSuchEmail
	}

	return &e, nil
}

func emailGetByUnsubscribeSecretHex(unsubscribeSecretHex models.HexID) (*models.Email, error) {
	row := svc.DB.QueryRow(
		fmt.Sprintf("select %s from emails where unsubscribesecrethex = $1;", emailsRowColumns),
		unsubscribeSecretHex)

	var e models.Email
	if err := emailsRowScan(row, &e); err != nil {
		// TODO: is this the only error?
		return nil, util.ErrorNoSuchUnsubscribeSecretHex
	}

	return &e, nil
}

func emailNotificationModerator(d *models.Domain, path string, title string, commenterHex models.HexID, commentHex models.HexID, html string, state models.CommentState) {
	if d.EmailNotificationPolicy == models.EmailNotificationPolicyNone ||
		d.EmailNotificationPolicy == models.EmailNotificationPolicyPendingDashModeration && state == models.CommentStateApproved {
		return
	}

	var commenterName string
	var commenterEmail strfmt.Email
	if commenterHex == "anonymous" {
		commenterName = "Anonymous"
	} else {
		c, err := commenterGetByHex(commenterHex)
		if err != nil {
			logger.Errorf("cannot get commenter to send email notification: %v", err)
			return
		}

		commenterName = c.Name
		commenterEmail = c.Email
	}

	kind := d.EmailNotificationPolicy
	if state != "approved" {
		kind = "pending-moderation"
	}

	for _, m := range d.Moderators {
		// Do not email the commenting moderator their own comment.
		if commenterHex != "anonymous" && m.Email == commenterEmail {
			continue
		}

		e, err := emailGet(m.Email)
		if err != nil {
			// No such email.
			continue
		}

		if !e.SendModeratorNotifications {
			continue
		}

		row := svc.DB.QueryRow("select name from commenters where email = $1;", m.Email)
		var name string
		if err := row.Scan(&name); err != nil {
			// The moderator has probably not created a commenter account.
			// We should only send emails to people who signed up, so skip.
			continue
		}

		if err := mail.SMTPEmailNotification(string(m.Email), name, string(kind), d.Domain, path, commentHex, commenterName, title, html, e.UnsubscribeSecretHex); err != nil {
			logger.Errorf("error sending email to %s: %v", m.Email, err)
			continue
		}
	}
}

func emailNotificationReply(d *models.Domain, path string, title string, commenterHex models.HexID, commentHex models.HexID, html string, parentHex models.HexID, state models.CommentState) {
	// No reply notifications for root comments.
	if parentHex == "root" {
		return
	}

	// No reply notification emails for unapproved comments.
	if state != models.CommentStateApproved {
		return
	}

	statement := `select commenterHex from comments where commentHex = $1;`
	row := svc.DB.QueryRow(statement, parentHex)

	var parentCommenterHex models.HexID
	err := row.Scan(&parentCommenterHex)
	if err != nil {
		logger.Errorf("cannot scan commenterHex and parentCommenterHex: %v", err)
		return
	}

	// No reply notification emails for anonymous users.
	if parentCommenterHex == "anonymous" {
		return
	}

	// No reply notification email for self replies.
	if parentCommenterHex == commenterHex {
		return
	}

	pc, err := commenterGetByHex(parentCommenterHex)
	if err != nil {
		logger.Errorf("cannot get commenter to send email notification: %v", err)
		return
	}

	var commenterName string
	if commenterHex == "anonymous" {
		commenterName = "Anonymous"
	} else {
		c, err := commenterGetByHex(commenterHex)
		if err != nil {
			logger.Errorf("cannot get commenter to send email notification: %v", err)
			return
		}
		commenterName = c.Name
	}

	epc, err := emailGet(pc.Email)
	if err != nil {
		// No such email.
		return
	}

	if !epc.SendReplyNotifications {
		return
	}

	_ = mail.SMTPEmailNotification(string(pc.Email), pc.Name, "reply", d.Domain, path, commentHex, commenterName, title, html, epc.UnsubscribeSecretHex)
}

func emailNotificationNew(d *models.Domain, path string, commenterHex models.HexID, commentHex models.HexID, html string, parentHex models.HexID, state models.CommentState) {
	p, err := pageGet(d.Domain, path)
	if err != nil {
		logger.Errorf("cannot get page to send email notification: %v", err)
		return
	}

	if p.Title == "" {
		p.Title, err = pageTitleUpdate(d.Domain, path)
		if err != nil {
			// Not being able to update a page title isn't serious enough to skip an email notification
			p.Title = d.Domain
		}
	}

	emailNotificationModerator(d, path, p.Title, commenterHex, commentHex, html, state)
	emailNotificationReply(d, path, p.Title, commenterHex, commentHex, html, parentHex, state)
}

func emailsRowScan(s util.Scanner, e *models.Email) error {
	return s.Scan(
		&e.Email,
		&e.UnsubscribeSecretHex,
		&e.LastEmailNotificationDate,
		&e.SendReplyNotifications,
		&e.SendModeratorNotifications,
	)
}
