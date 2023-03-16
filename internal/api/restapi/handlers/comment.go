package handlers

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/markbates/goth"
	"gitlab.com/comentario/comentario/internal/api/models"
	"gitlab.com/comentario/comentario/internal/api/restapi/operations"
	"gitlab.com/comentario/comentario/internal/data"
	"gitlab.com/comentario/comentario/internal/svc"
	"gitlab.com/comentario/comentario/internal/util"
	"time"
)

func CommentApprove(params operations.CommentApproveParams) middleware.Responder {
	// Find the commenter
	commenter, err := svc.TheUserService.FindCommenterByToken(*params.Body.CommenterToken)
	if err != nil {
		return respServiceError(err)
	}

	// Fetch the comment
	comment, err := svc.TheCommentService.FindByHexID(*params.Body.CommentHex)
	if err != nil {
		return respServiceError(err)
	}

	// Verify the user is a domain moderator
	if isModerator, err := svc.TheDomainService.IsDomainModerator(comment.Domain, commenter.Email); err != nil {
		return respServiceError(err)
	} else if !isModerator {
		return operations.NewGenericForbidden()
	}

	// Update the comment's state in the database
	if err = svc.TheCommentService.Approve(*params.Body.CommentHex); err != nil {
		return respServiceError(err)
	}

	// Succeeded
	return operations.NewCommentApproveOK().WithPayload(&models.APIResponseBase{Success: true})
}

func CommentCount(params operations.CommentCountParams) middleware.Responder {
	// Fetch comment counts
	cc, err := svc.ThePageService.CommentCountsByPath(*params.Body.Domain, params.Body.Paths)
	if err != nil {
		return respServiceError(err)
	}

	// Succeeded
	return operations.NewCommentCountOK().WithPayload(&operations.CommentCountOKBody{CommentCounts: cc})
}

func CommentDelete(params operations.CommentDeleteParams) middleware.Responder {
	// Find the commenter
	commenter, err := svc.TheUserService.FindCommenterByToken(*params.Body.CommenterToken)
	if err != nil {
		return respServiceError(err)
	}

	// Find the comment
	comment, err := svc.TheCommentService.FindByHexID(*params.Body.CommentHex)
	if err != nil {
		return respServiceError(err)
	}

	// If not deleting their own comment, the user must be a domain moderator
	if comment.CommenterHex != commenter.CommenterHexID() {
		if isModerator, err := svc.TheDomainService.IsDomainModerator(comment.Domain, commenter.Email); err != nil {
			return respServiceError(err)
		} else if !isModerator {
			return operations.NewGenericForbidden()
		}
	}

	// Mark the comment deleted
	if err = svc.TheCommentService.MarkDeleted(comment.CommentHex, commenter.CommenterHexID()); err != nil {
		return respServiceError(err)
	}

	// Succeeded
	return operations.NewCommentDeleteOK().WithPayload(&models.APIResponseBase{Success: true})
}

func CommentEdit(params operations.CommentEditParams) middleware.Responder {
	// Find the commenter
	commenter, err := svc.TheUserService.FindCommenterByToken(*params.Body.CommenterToken)
	if err != nil {
		return respServiceError(err)
	}

	// Find the existing comment
	comment, err := svc.TheCommentService.FindByHexID(*params.Body.CommentHex)
	if err != nil {
		return respServiceError(err)
	}

	// If not updating their own comment, the user must be a domain moderator
	if comment.CommenterHex != commenter.CommenterHexID() {
		if isModerator, err := svc.TheDomainService.IsDomainModerator(comment.Domain, commenter.Email); err != nil {
			return respServiceError(err)
		} else if !isModerator {
			return operations.NewGenericForbidden()
		}
	}

	// Render the comment into HTML
	markdown := swag.StringValue(params.Body.Markdown)
	html := util.MarkdownToHTML(markdown)

	// Persist the edits in the database
	if err := svc.TheCommentService.UpdateText(*params.Body.CommentHex, markdown, html); err != nil {
		return operations.NewCommentEditOK().WithPayload(&operations.CommentEditOKBody{Message: err.Error()})
	}

	// Succeeded
	return operations.NewCommentEditOK().WithPayload(&operations.CommentEditOKBody{
		HTML:    html,
		Success: true,
	})
}

func CommentList(params operations.CommentListParams) middleware.Responder {
	// Fetch the domain
	domain, err := svc.TheDomainService.FindByName(*params.Body.Domain)
	if err != nil {
		return respServiceError(err)
	}

	// Prepare a map of configured identity providers: federated ones should only be enabled when configured
	idps := domain.Idps.Clone()
	for idp, gothIdP := range util.FederatedIdProviders {
		idps[idp] = idps[idp] && goth.GetProviders()[gothIdP] != nil
	}

	// Fetch the page
	page, err := svc.ThePageService.FindByDomainPath(domain.Domain, params.Body.Path)
	if err != nil {
		return respServiceError(err)
	}

	// If it isn't an anonymous token, try to find the related Commenter
	var commenter *data.UserCommenter
	commenterHex := data.AnonymousCommenterHexID
	if *params.Body.CommenterToken != data.AnonymousCommenterHexID {
		if commenter, err = svc.TheUserService.FindCommenterByToken(*params.Body.CommenterToken); err != nil {
			return respServiceError(err)
		}
		commenterHex = commenter.CommenterHexID()
	}

	// Make a map of moderator emails, also figure out if the user is a moderator self
	isModerator := false
	moderatorEmailMap := map[strfmt.Email]bool{}
	for _, mod := range domain.Moderators {
		moderatorEmailMap[mod.Email] = true
		if commenter != nil && string(mod.Email) == commenter.Email {
			isModerator = true
		}
	}

	// Fetch comment list
	comments, commenters, err := svc.TheCommentService.ListByDomainPath(commenterHex, domain.Domain, params.Body.Path, isModerator)
	if err != nil {
		return operations.NewCommentListOK().WithPayload(&operations.CommentListOKBody{Message: err.Error()})
	}

	// Update each commenter
	for _, cr := range commenters {
		// Set the IsModerator flag to true for domain moderators
		if moderatorEmailMap[cr.Email] {
			cr.IsModerator = true
		}

		// Wipe out the email (to omit them in the response)
		cr.Email = ""
	}

	// Register a view in domain statistics, ignoring any error
	_ = svc.TheDomainService.RegisterView(domain.Domain, commenterHex)

	// Succeeded
	return operations.NewCommentListOK().WithPayload(&operations.CommentListOKBody{
		Attributes:            page,
		Commenters:            commenters,
		Comments:              comments,
		ConfiguredOauths:      idps,
		DefaultSortPolicy:     domain.DefaultSortPolicy,
		Domain:                domain.Domain,
		IsFrozen:              domain.State == models.DomainStateFrozen,
		IsModerator:           isModerator,
		RequireIdentification: domain.RequireIdentification,
		RequireModeration:     domain.RequireModeration,
		Success:               true,
	})
}

func CommentNew(params operations.CommentNewParams) middleware.Responder {
	// Fetch the domain
	domain, err := svc.TheDomainService.FindByName(*params.Body.Domain)
	if err != nil {
		return respServiceError(err)
	}

	// Verify the domain isn't frozen
	if domain.State == models.DomainStateFrozen {
		return operations.NewGenericBadRequest().WithPayload(&operations.GenericBadRequestBody{Details: util.ErrorDomainFrozen.Error()})
	}

	// Verify that either the domain allows anonymous commenting or the user is authenticated
	if domain.RequireIdentification && *params.Body.CommenterToken == data.AnonymousCommenterHexID {
		return operations.NewGenericUnauthorized()
	}

	commenterHex := data.AnonymousCommenterHexID
	commenterEmail := strfmt.Email("")
	commenterName := "Anonymous"
	commenterWebsite := ""
	var isModerator bool
	if *params.Body.CommenterToken != data.AnonymousCommenterHexID {
		// Find the commenter
		commenter, err := svc.TheUserService.FindCommenterByToken(*params.Body.CommenterToken)
		if err != nil {
			return respServiceError(err)
		}
		commenterHex = commenter.CommenterHexID()
		commenterEmail = strfmt.Email(commenter.Email)
		commenterName = commenter.Name
		commenterWebsite = commenter.WebsiteURL
		for _, mod := range domain.Moderators {
			if mod.Email == commenterEmail {
				isModerator = true
				break
			}
		}
	}

	// Determine comment state
	var state models.CommentState
	if isModerator {
		state = models.CommentStateApproved
	} else if domain.RequireModeration || commenterHex == data.AnonymousCommenterHexID && domain.ModerateAllAnonymous {
		state = models.CommentStateUnapproved
	} else if domain.AutoSpamFilter && checkForSpam(*params.Body.Domain, util.UserIP(params.HTTPRequest), util.UserAgent(params.HTTPRequest), commenterName, string(commenterEmail), commenterWebsite, *params.Body.Markdown) {
		state = models.CommentStateFlagged
	} else {
		state = models.CommentStateApproved
	}

	// Persist a new comment record
	comment, err := svc.TheCommentService.Create(
		commenterHex,
		domain.Domain,
		params.Body.Path,
		*params.Body.Markdown,
		*params.Body.ParentHex,
		state,
		strfmt.DateTime(time.Now().UTC()))
	if err != nil {
		return respServiceError(err)
	}

	// Send out an email notification
	go emailNotificationNew(domain, comment)

	// Succeeded
	return operations.NewCommentNewOK().WithPayload(&operations.CommentNewOKBody{
		CommentHex: comment.CommentHex,
		HTML:       comment.HTML,
		State:      state,
		Success:    true,
	})
}

func CommentVote(params operations.CommentVoteParams) middleware.Responder {
	// Find the commenter
	commenter, err := svc.TheUserService.FindCommenterByToken(*params.Body.CommenterToken)
	if err != nil {
		return respServiceError(err)
	}

	// Calculate the direction
	direction := 0
	if *params.Body.Direction > 0 {
		direction = 1
	} else if *params.Body.Direction < 0 {
		direction = -1
	}

	// Find the comment
	comment, err := svc.TheCommentService.FindByHexID(*params.Body.CommentHex)
	if err != nil {
		return respServiceError(err)
	}

	// Make sure the commenter is not voting for their own comment
	if comment.CommenterHex == commenter.CommenterHexID() {
		return operations.NewGenericForbidden().WithPayload(&operations.GenericForbiddenBody{Details: util.ErrorSelfVote.Error()})
	}

	// Update the vote in the database
	if err := svc.TheVoteService.SetVote(comment.CommentHex, commenter.CommenterHexID(), direction); err != nil {
		return respServiceError(err)
	}

	// Succeeded
	return operations.NewCommentVoteOK().WithPayload(&models.APIResponseBase{Success: true})
}
