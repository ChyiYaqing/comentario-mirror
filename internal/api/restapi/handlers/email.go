package handlers

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"gitlab.com/comentario/comentario/internal/api/models"
	"gitlab.com/comentario/comentario/internal/api/restapi/operations"
	"gitlab.com/comentario/comentario/internal/data"
	"gitlab.com/comentario/comentario/internal/svc"
	"gitlab.com/comentario/comentario/internal/util"
	"time"
)

func EmailGet(params operations.EmailGetParams) middleware.Responder {
	// Fetch the email by its unsubscribe token
	email, err := svc.TheEmailService.FindByUnsubscribeToken(*params.Body.UnsubscribeSecretHex)
	if err != nil {
		return serviceErrorResponder(err)
	}

	// Succeeded
	return operations.NewEmailGetOK().WithPayload(&operations.EmailGetOKBody{
		Email:   email,
		Success: true,
	})
}

func EmailModerate(params operations.EmailModerateParams) middleware.Responder {
	// Find the comment
	comment, err := svc.TheCommentService.FindByHexID(models.HexID(params.CommentHex))
	if err != nil {
		return serviceErrorResponder(err)
	}

	// If the comment is already deleted
	if comment.Deleted {
		return operations.NewGenericBadRequest().WithPayload(&operations.GenericBadRequestBody{Details: "Comment has already been deleted"})
	}

	// Fetch the email by its unsubscribe token
	email, err := svc.TheEmailService.FindByUnsubscribeToken(models.HexID(params.UnsubscribeSecretHex))
	if err != nil {
		return serviceErrorResponder(err)
	}

	// Verify the user is a moderator
	if isModerator, err := svc.TheDomainService.IsDomainModerator(comment.Domain, string(email.Email)); err != nil {
		return serviceErrorResponder(err)
	} else if !isModerator {
		return operations.NewGenericForbidden()
	}

	// TODO Do not use commenterGetByEmail here because we don't know which provider should be used. This was poor design on
	// multiple fronts on my part, but let's deal with that later. For now, it suffices to match the deleter/approver
	// with any account owned by the same email
	row := svc.DB.QueryRow("select commenterHex from commenters where email = $1;", email.Email)
	var commenterHex models.CommenterHexID
	if err = row.Scan(&commenterHex); err != nil {
		logger.Errorf("cannot retrieve commenterHex by email %q: %v", email.Email, err)
		return operations.NewGenericInternalServerError()
	}

	switch params.Action {
	case "approve":
		if err := svc.TheCommentService.Approve(comment.CommentHex); err != nil {
			return serviceErrorResponder(err)
		}
	case "delete":
		if err := svc.TheCommentService.MarkDeleted(comment.CommentHex, commenterHex); err != nil {
			return serviceErrorResponder(err)
		}
	default:
		return operations.NewGenericBadRequest().WithPayload(&operations.GenericBadRequestBody{Details: util.ErrorInvalidAction.Error()})
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

	err = svc.DB.Exec(
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
	// Update the email record
	err := svc.TheEmailService.UpdateByEmailToken(
		string(params.Body.Email.Email),
		params.Body.Email.UnsubscribeSecretHex,
		params.Body.Email.SendReplyNotifications,
		params.Body.Email.SendModeratorNotifications)
	if err != nil {
		return serviceErrorResponder(err)
	}

	// Succeeded
	return operations.NewEmailUpdateOK().WithPayload(&models.APIResponseBase{Success: true})
}

func emailNotificationModerator(d *models.Domain, path string, title string, commenterHex models.CommenterHexID, commentHex models.HexID, html string, state models.CommentState) {
	commenterName := "Anonymous"
	var commenterEmail strfmt.Email
	if commenterHex != data.AnonymousCommenterHexID {
		if commenter, err := svc.TheUserService.FindCommenterByID(commenterHex); err != nil {
			return
		} else {
			commenterName = commenter.Name
			commenterEmail = strfmt.Email(commenter.Email)
		}
	}

	kind := d.EmailNotificationPolicy
	if state != models.CommentStateApproved {
		kind = models.EmailNotificationPolicyPendingDashModeration
	}

	for _, m := range d.Moderators {
		// Do not email the commenting moderator their own comment.
		if commenterEmail != "" && m.Email == commenterEmail {
			continue
		}

		// Try to fetch the moderator's email to check whether the notifications are enabled
		modEmail, err := svc.TheEmailService.FindByEmail(string(m.Email))
		if err != nil || !modEmail.SendModeratorNotifications {
			continue
		}

		row := svc.DB.QueryRow("select name from commenters where email = $1;", m.Email)
		var name string
		if err := row.Scan(&name); err != nil {
			// The moderator has probably not created a commenter account.
			// We should only send emails to people who signed up, so skip.
			continue
		}

		// Send a notification (ignore errors)
		_ = svc.TheMailService.SendCommentNotification(
			string(m.Email),
			string(kind),
			d.Domain,
			path,
			commenterName,
			title,
			html,
			commentHex,
			modEmail.UnsubscribeSecretHex)
	}
}

func emailNotificationReply(d *models.Domain, path string, title string, commenterHex models.CommenterHexID, commentHex models.HexID, html string, parentHex models.ParentHexID) {
	row := svc.DB.QueryRow("select commenterHex from comments where commentHex = $1;", parentHex)
	var parentCommenterHex models.CommenterHexID
	err := row.Scan(&parentCommenterHex)
	if err != nil {
		logger.Errorf("cannot scan commenterHex and parentCommenterHex: %v", err)
		return
	}

	// No reply notification emails for anonymous users and self replies
	if parentCommenterHex == data.AnonymousCommenterHexID || parentCommenterHex == commenterHex {
		return
	}

	// Find the parent commenter
	parentCommenter, err := svc.TheUserService.FindCommenterByID(parentCommenterHex)
	if err != nil {
		return
	}

	commenterName := "Anonymous"
	if commenterHex != data.AnonymousCommenterHexID {
		if commenter, err := svc.TheUserService.FindCommenterByID(commenterHex); err != nil {
			return
		} else {
			commenterName = commenter.Name
		}
	}

	// Fetch the parent commenter's email
	parentEmail, err := svc.TheEmailService.FindByEmail(parentCommenter.Email)
	if err != nil {
		// No valid email, ignore
		return
	}

	// Queue a notification, if the notifications are enabled for this email (ignore errors)
	if parentEmail.SendReplyNotifications {
		_ = svc.TheMailService.SendCommentNotification(
			parentCommenter.Email,
			"reply",
			d.Domain,
			path,
			commenterName,
			title,
			html,
			commentHex,
			parentEmail.UnsubscribeSecretHex)
	}
}

func emailNotificationNew(d *models.Domain, c *models.Comment) {
	// Fetch the page
	page, err := svc.ThePageService.FindByDomainPath(d.Domain, c.URL)
	if err != nil {
		logger.Errorf("cannot get page to send email notification: %v", err)
		return
	}

	// If the page has no title, try to fetch it
	if page.Title == "" {
		if page.Title, err = svc.ThePageService.UpdateTitleByDomainPath(d.Domain, c.URL); err != nil {
			// Failed, just use the domain name
			page.Title = d.Domain
		}
	}

	// Send an email notification to moderators, if we notify about every comment or comments pending moderation and
	// the comment isn't approved yet
	if d.EmailNotificationPolicy == models.EmailNotificationPolicyAll || d.EmailNotificationPolicy == models.EmailNotificationPolicyPendingDashModeration && c.State != models.CommentStateApproved {
		emailNotificationModerator(d, c.URL, page.Title, c.CommenterHex, c.CommentHex, c.HTML, c.State)
	}

	// If it's a reply and the comment is approved, send out a reply notifications
	if c.ParentHex != data.RootParentHexID && c.State == models.CommentStateApproved {
		emailNotificationReply(d, c.URL, page.Title, c.CommenterHex, c.CommentHex, c.HTML, c.ParentHex)
	}
}
