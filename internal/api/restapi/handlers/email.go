package handlers

import (
	"github.com/go-openapi/runtime/middleware"
	"gitlab.com/comentario/comentario/internal/api/models"
	"gitlab.com/comentario/comentario/internal/api/restapi/operations"
	"gitlab.com/comentario/comentario/internal/data"
	"gitlab.com/comentario/comentario/internal/svc"
	"gitlab.com/comentario/comentario/internal/util"
)

func EmailGet(params operations.EmailGetParams) middleware.Responder {
	// Fetch the email by its unsubscribe token
	email, err := svc.TheEmailService.FindByUnsubscribeToken(*params.Body.UnsubscribeSecretHex)
	if err != nil {
		return respServiceError(err)
	}

	// Succeeded
	return operations.NewEmailGetOK().WithPayload(&operations.EmailGetOKBody{Email: email})
}

func EmailModerate(params operations.EmailModerateParams) middleware.Responder {
	// Find the comment
	comment, err := svc.TheCommentService.FindByHexID(models.HexID(params.CommentHex))
	if err != nil {
		return respServiceError(err)
	}

	// Verify the comment isn't deleted yet
	if comment.Deleted {
		return respBadRequest(util.ErrorCommentDeleted)
	}

	// Fetch the email by its unsubscribe token
	email, err := svc.TheEmailService.FindByUnsubscribeToken(models.HexID(params.UnsubscribeSecretHex))
	if err != nil {
		return respServiceError(err)
	}

	// Verify the user is a domain moderator
	if r := Verifier.UserIsDomainModerator(string(email.Email), comment.Domain); r != nil {
		return r
	}

	// TODO this must be changed to using hex ID or IdP
	// Find (any) commenter with that email
	commenter, err := svc.TheUserService.FindCommenterByEmail(string(email.Email))
	if err != nil {
		return respServiceError(err)
	}

	// Perform the appropriate action
	switch params.Action {
	case "approve":
		if err := svc.TheCommentService.Approve(comment.CommentHex); err != nil {
			return respServiceError(err)
		}
	case "delete":
		if err := svc.TheCommentService.MarkDeleted(comment.CommentHex, commenter.HexID); err != nil {
			return respServiceError(err)
		}
	default:
		return respBadRequest(util.ErrorInvalidAction)
	}

	// Succeeded
	// TODO redirect to a proper page instead of letting the user see JSON response
	return operations.NewEmailModerateNoContent()
}

func EmailUpdate(params operations.EmailUpdateParams) middleware.Responder {
	// Update the email record
	err := svc.TheEmailService.UpdateByEmailToken(
		string(params.Body.Email.Email),
		params.Body.Email.UnsubscribeSecretHex,
		params.Body.Email.SendReplyNotifications,
		params.Body.Email.SendModeratorNotifications)
	if err != nil {
		return respServiceError(err)
	}

	// Succeeded
	return operations.NewEmailUpdateNoContent()
}

func emailNotificationModerator(d *models.Domain, path string, title string, commenterHex models.HexID, commentHex models.HexID, html string, state models.CommentState) {
	// Find the related commenter
	commenter := &data.AnonymousCommenter
	if commenterHex != data.AnonymousCommenter.HexID {
		var err error
		if commenter, err = svc.TheUserService.FindCommenterByID(commenterHex); err != nil {
			// Failed to retrieve, give up
			return
		}
	}

	// Determine notification kind
	kind := d.EmailNotificationPolicy
	if state != models.CommentStateApproved {
		kind = models.EmailNotificationPolicyPendingDashModeration
	}

	for _, m := range d.Moderators {
		// Do not email the commenting moderator their own comment
		if commenter.Email != "" && string(m.Email) == commenter.Email {
			continue
		}

		// Try to fetch the moderator's email to check whether the notifications are enabled
		modEmail, err := svc.TheEmailService.FindByEmail(string(m.Email))
		if err != nil || !modEmail.SendModeratorNotifications {
			continue
		}

		// Send a notification (ignore errors)
		_ = svc.TheMailService.SendCommentNotification(
			string(m.Email),
			string(kind),
			d.Domain,
			path,
			commenter.Name,
			title,
			html,
			commentHex,
			modEmail.UnsubscribeSecretHex)
	}
}

func emailNotificationReply(d *models.Domain, path string, title string, commenterHex, commentHex, parentHex models.HexID, html string) {
	// Fetch the parent comment
	parentComment, err := svc.TheCommentService.FindByHexID(parentHex)
	if err != nil {
		return
	}

	// No reply notification emails for anonymous users and self replies
	if parentComment.CommenterHex == data.AnonymousCommenter.HexID || parentComment.CommenterHex == commenterHex {
		return
	}

	// Find the parent commenter
	parentCommenter, err := svc.TheUserService.FindCommenterByID(parentComment.CommenterHex)
	if err != nil {
		return
	}

	// Find the commenter for the comment in question
	commenter := &data.AnonymousCommenter
	if commenterHex != data.AnonymousCommenter.HexID {
		if commenter, err = svc.TheUserService.FindCommenterByID(commenterHex); err != nil {
			return
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
			commenter.Name,
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
		emailNotificationReply(d, c.URL, page.Title, c.CommenterHex, c.CommentHex, models.HexID(c.ParentHex), c.HTML)
	}
}
