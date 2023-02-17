package handlers

import (
	"fmt"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"gitlab.com/comentario/comentario/internal/api/models"
	"gitlab.com/comentario/comentario/internal/api/restapi/operations"
	"gitlab.com/comentario/comentario/internal/config"
	"gitlab.com/comentario/comentario/internal/svc"
	"gitlab.com/comentario/comentario/internal/util"
	"html/template"
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

func EmailModerate(params operations.EmailModerateParams) middleware.Responder {
	row := svc.DB.QueryRow("select domain, deleted from comments where commentHex = $1;", params.CommentHex)

	var domain string
	var deleted bool
	if err := row.Scan(&domain, &deleted); err != nil {
		// TODO: is this the only error?
		return operations.NewGenericBadRequest().WithPayload(&operations.GenericBadRequestBody{Details: "No such comment found (perhaps it has been deleted?)"})
	}
	if deleted {
		return operations.NewGenericBadRequest().WithPayload(&operations.GenericBadRequestBody{Details: "Comment has already been deleted"})
	}

	e, err := emailGetByUnsubscribeSecretHex(models.HexID(params.UnsubscribeSecretHex))
	if err != nil {
		return operations.NewGenericBadRequest().WithPayload(&operations.GenericBadRequestBody{Details: err.Error()})
	}

	isModerator, err := isDomainModerator(domain, e.Email)
	if err != nil {
		logger.Errorf("error checking if %s is a moderator: %v", e.Email, err)
		return operations.NewGenericInternalServerError()
	}

	if !isModerator {
		return operations.NewGenericBadRequest().WithPayload(&operations.GenericBadRequestBody{Details: "Not a moderator of that domain"})
	}

	// Do not use commenterGetByEmail here because we don't know which provider should be used. This was poor design on
	// multiple fronts on my part, but let's deal with that later. For now, it suffices to match the deleter/approver
	// with any account owned by the same email
	row = svc.DB.QueryRow("select commenterHex from commenters where email = $1;", e.Email)

	var commenterHex models.HexID
	if err = row.Scan(&commenterHex); err != nil {
		logger.Errorf("cannot retrieve commenterHex by email %q: %v", e.Email, err)
		return operations.NewGenericInternalServerError()
	}

	switch params.Action {
	case "approve":
		err = commentApprove(models.HexID(params.CommentHex))
	case "delete":
		err = commentDelete(models.HexID(params.CommentHex), commenterHex)
	default:
		return operations.NewGenericBadRequest().WithPayload(&operations.GenericBadRequestBody{Details: util.ErrorInvalidAction.Error()})
	}

	if err != nil {
		return operations.NewGenericBadRequest().WithPayload(&operations.GenericBadRequestBody{Details: err.Error()})
	}

	// Succeeded
	// TODO redirect to a proper page instead of letting the user see JSON response
	return operations.NewEmailModerateOK().WithPayload(&models.APIResponseBase{Success: true})
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

func EmailUpdate(params operations.EmailUpdateParams) middleware.Responder {
	if err := emailUpdate(params.Body.Email); err != nil {
		return operations.NewEmailUpdateOK().WithPayload(&models.APIResponseBase{Message: err.Error()})
	}

	// Succeeded
	return operations.NewEmailUpdateOK().WithPayload(&models.APIResponseBase{Success: true})
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

		emailSendNotification(m.Email, string(kind), d.Domain, path, commenterName, title, html, commentHex, e.UnsubscribeSecretHex)
	}
}

func emailNotificationReply(d *models.Domain, path string, title string, commenterHex models.HexID, commentHex models.HexID, html string, parentHex models.ParentHexID) {
	row := svc.DB.QueryRow("select commenterHex from comments where commentHex = $1;", parentHex)
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

	if epc.SendReplyNotifications {
		emailSendNotification(pc.Email, "reply", d.Domain, path, commenterName, title, html, commentHex, epc.UnsubscribeSecretHex)
	}
}

func emailNotificationNew(d *models.Domain, path string, commenterHex models.HexID, commentHex models.HexID, html string, parentHex models.ParentHexID, state models.CommentState) {
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

	// Send an email notification to moderators, if we notify about every comment or comments pending moderation and
	// the comment isn't approved yet
	if d.EmailNotificationPolicy == models.EmailNotificationPolicyAll || d.EmailNotificationPolicy == models.EmailNotificationPolicyPendingDashModeration && state != models.CommentStateApproved {
		emailNotificationModerator(d, path, p.Title, commenterHex, commentHex, html, state)
	}

	// If it's a reply and the comment is approved, send out a reply notifications
	if parentHex != RootParentHexID && state == models.CommentStateApproved {
		emailNotificationReply(d, path, p.Title, commenterHex, commentHex, html, parentHex)
	}
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

func emailSendNotification(recipientEmail strfmt.Email, kind string, domain, path, commenterName, title, html string, commentHex, unsubscribeSecretHex models.HexID) {
	_ = svc.TheEmailService.SendFromTemplate(
		"",
		string(recipientEmail),
		"Comentario: "+title,
		"email-notification.gohtml",
		map[string]any{
			"Kind":          kind,
			"Title":         title,
			"Domain":        domain,
			"Path":          path,
			"CommentHex":    commentHex,
			"CommenterName": commenterName,
			"HTML":          template.HTML(html),
			"ApproveURL": config.URLForAPI(
				"email/moderate",
				map[string]string{"action": "approve", "commentHex": string(commentHex), "unsubscribeSecretHex": string(unsubscribeSecretHex)}),
			"DeleteURL": config.URLForAPI(
				"email/moderate",
				map[string]string{"action": "delete", "commentHex": string(commentHex), "unsubscribeSecretHex": string(unsubscribeSecretHex)}),
			"UnsubscribeURL": config.URLFor(
				"unsubscribe",
				map[string]string{"unsubscribeSecretHex": string(unsubscribeSecretHex)}),
		})
}

func emailUpdate(e *models.Email) error {
	_, err := svc.DB.Exec(
		"update emails set sendReplyNotifications = $3, sendModeratorNotifications = $4 where email = $1 and unsubscribeSecretHex = $2;",
		e.Email,
		e.UnsubscribeSecretHex,
		e.SendReplyNotifications,
		e.SendModeratorNotifications)
	if err != nil {
		logger.Errorf("error updating email: %v", err)
		return util.ErrorInternal
	}

	return nil
}
