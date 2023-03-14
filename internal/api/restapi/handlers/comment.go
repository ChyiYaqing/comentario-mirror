package handlers

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/lib/pq"
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
		return serviceErrorResponder(err)
	}

	// Fetch the comment
	comment, err := svc.TheCommentService.FindByHexID(*params.Body.CommentHex)
	if err != nil {
		return serviceErrorResponder(err)
	}

	// Verify the user is a domain moderator
	if isModerator, err := isDomainModerator(comment.Domain, strfmt.Email(commenter.Email)); err != nil {
		return operations.NewCommentApproveOK().WithPayload(&models.APIResponseBase{Message: err.Error()})
	} else if !isModerator {
		return operations.NewCommentApproveOK().WithPayload(&models.APIResponseBase{Message: util.ErrorNotModerator.Error()})
	}

	// Update the comment's state in the database
	if err = svc.TheCommentService.Approve(*params.Body.CommentHex); err != nil {
		return serviceErrorResponder(err)
	}

	// Succeeded
	return operations.NewCommentApproveOK().WithPayload(&models.APIResponseBase{Success: true})
}

func CommentCount(params operations.CommentCountParams) middleware.Responder {
	commentCounts, err := commentCount(*params.Body.Domain, params.Body.Paths)
	if err != nil {
		return operations.NewCommentCountOK().WithPayload(&operations.CommentCountOKBody{Message: err.Error()})
	}

	// Succeeded
	return operations.NewCommentCountOK().WithPayload(&operations.CommentCountOKBody{
		Success:       true,
		CommentCounts: commentCounts,
	})
}

func CommentDelete(params operations.CommentDeleteParams) middleware.Responder {
	// Find the commenter
	commenter, err := svc.TheUserService.FindCommenterByToken(*params.Body.CommenterToken)
	if err != nil {
		return serviceErrorResponder(err)
	}

	// Find the comment
	comment, err := svc.TheCommentService.FindByHexID(*params.Body.CommentHex)
	if err != nil {
		return operations.NewCommentDeleteOK().WithPayload(&models.APIResponseBase{Message: err.Error()})
	}

	// If not deleting their own comment, the user must be a domain moderator
	if comment.CommenterHex != commenter.CommenterHexID() {
		// Find the domain moderator
		if isModerator, err := isDomainModerator(comment.Domain, strfmt.Email(commenter.Email)); err != nil {
			return operations.NewCommentDeleteOK().WithPayload(&models.APIResponseBase{Message: err.Error()})
			// Check the commenter is a domain moderator
		} else if !isModerator {
			return operations.NewCommentDeleteOK().WithPayload(&models.APIResponseBase{Message: util.ErrorNotModerator.Error()})
		}
	}

	if err = commentDelete(*params.Body.CommentHex, commenter.HexID); err != nil {
		return operations.NewCommentDeleteOK().WithPayload(&models.APIResponseBase{Message: err.Error()})
	}

	// Succeeded
	return operations.NewCommentDeleteOK().WithPayload(&models.APIResponseBase{Success: true})
}

func CommentEdit(params operations.CommentEditParams) middleware.Responder {
	// Find the commenter
	commenter, err := svc.TheUserService.FindCommenterByToken(*params.Body.CommenterToken)
	if err != nil {
		return serviceErrorResponder(err)
	}

	// Find the existing comment
	comment, err := svc.TheCommentService.FindByHexID(*params.Body.CommentHex)
	if err != nil {
		return serviceErrorResponder(err)
	}

	// If not updating their own comment, the user must be a domain moderator
	if comment.CommenterHex != commenter.CommenterHexID() {
		// Find the domain moderator
		if isModerator, err := isDomainModerator(comment.Domain, strfmt.Email(commenter.Email)); err != nil {
			return operations.NewCommentEditOK().WithPayload(&operations.CommentEditOKBody{Message: err.Error()})
			// Check the commenter is a domain moderator
		} else if !isModerator {
			return operations.NewCommentEditOK().WithPayload(&operations.CommentEditOKBody{Message: util.ErrorNotModerator.Error()})
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
	domainName := *params.Body.Domain
	domain, err := domainGet(domainName)
	if err != nil {
		return operations.NewCommentListOK().WithPayload(&operations.CommentListOKBody{Message: err.Error()})
	}

	// Fetch the page
	page, err := svc.ThePageService.FindByDomainPath(domainName, params.Body.Path)
	if err != nil {
		return serviceErrorResponder(err)
	}

	// If it isn't an anonymous token, try to find the related Commenter
	var commenter *data.UserCommenter
	if *params.Body.CommenterToken != data.AnonymousCommenterHexID {
		if commenter, err = svc.TheUserService.FindCommenterByToken(*params.Body.CommenterToken); err != nil {
			return serviceErrorResponder(err)
		}
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

	// Register a view in domain statistics
	domainViewRecord(domainName, commenter)

	// Fetch comment list
	comments, commenters, err := commentList(commenter, domainName, params.Body.Path, isModerator)
	if err != nil {
		return operations.NewCommentListOK().WithPayload(&operations.CommentListOKBody{Message: err.Error()})
	}

	_commenters := map[models.CommenterHexID]*models.Commenter{}
	for ch, cr := range commenters {
		if moderatorEmailMap[cr.Email] {
			cr.IsModerator = true
		}
		cr.Email = ""
		_commenters[ch] = cr
	}

	// Prepare a map of configured identity providers: federated ones should only be enabled when configured
	idps := domain.Idps.Clone()
	for idp, gothIdP := range util.FederatedIdProviders {
		idps[idp] = idps[idp] && goth.GetProviders()[gothIdP] != nil
	}

	return operations.NewCommentListOK().WithPayload(&operations.CommentListOKBody{
		Attributes:            page,
		Commenters:            _commenters,
		Comments:              comments,
		ConfiguredOauths:      idps,
		DefaultSortPolicy:     domain.DefaultSortPolicy,
		Domain:                domainName,
		IsFrozen:              domain.State == models.DomainStateFrozen,
		IsModerator:           isModerator,
		RequireIdentification: domain.RequireIdentification,
		RequireModeration:     domain.RequireModeration,
		Success:               true,
	})
}

func CommentNew(params operations.CommentNewParams) middleware.Responder {
	domainName := *params.Body.Domain
	domain, err := domainGet(domainName)
	if err != nil {
		return operations.NewCommentNewOK().WithPayload(&operations.CommentNewOKBody{Message: err.Error()})
	}

	if domain.State == "frozen" {
		return operations.NewCommentNewOK().WithPayload(&operations.CommentNewOKBody{Message: util.ErrorDomainFrozen.Error()})
	}

	if domain.RequireIdentification && *params.Body.CommenterToken == data.AnonymousCommenterHexID {
		return operations.NewCommentNewOK().WithPayload(&operations.CommentNewOKBody{Message: util.ErrorNotAuthorised.Error()})
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
			return serviceErrorResponder(err)
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
	comment, err := svc.TheCommentService.Create(commenterHex, domainName, params.Body.Path, *params.Body.Markdown, *params.Body.ParentHex, state, strfmt.DateTime(time.Now().UTC()))
	if err != nil {
		return serviceErrorResponder(err)
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
		return serviceErrorResponder(err)
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
		return serviceErrorResponder(err)
	}

	// Make sure the commenter is not voting for their own comment
	if comment.CommenterHex == commenter.CommenterHexID() {
		return operations.NewGenericForbidden().WithPayload(&operations.GenericForbiddenBody{Details: util.ErrorSelfVote.Error()})
	}

	// Update the vote in the database
	if err := svc.TheVoteService.SetVote(comment.CommentHex, commenter.CommenterHexID(), direction); err != nil {
		return serviceErrorResponder(err)
	}

	// Succeeded
	return operations.NewCommentVoteOK().WithPayload(&models.APIResponseBase{Success: true})
}

func commentCount(domain string, paths []string) (map[string]int, error) {
	commentCounts := map[string]int{}

	if domain == "" {
		return nil, util.ErrorMissingField
	}

	if len(paths) == 0 {
		return nil, util.ErrorEmptyPaths
	}

	rows, err := svc.DB.Query(
		"select path, commentCount from pages where domain = $1 and path = any($2);",
		domain,
		pq.Array(paths))
	if err != nil {
		logger.Errorf("cannot get comments: %v", err)
		return nil, util.ErrorInternal
	}
	defer rows.Close()

	for rows.Next() {
		var path string
		var commentCount int
		if err = rows.Scan(&path, &commentCount); err != nil {
			logger.Errorf("cannot scan path and commentCount: %v", err)
			return nil, util.ErrorInternal
		}

		commentCounts[path] = commentCount
	}

	return commentCounts, nil
}

func commentDelete(commentHex models.HexID, deleterHex models.HexID) error {
	if commentHex == "" || deleterHex == "" {
		return util.ErrorMissingField
	}

	err := svc.DB.Exec(
		"update comments "+
			"set deleted = true, markdown = '[deleted]', html = '[deleted]', commenterHex = 'anonymous', deleterHex = $2, deletionDate = $3 "+
			"where commentHex = $1;",
		commentHex,
		deleterHex,
		time.Now().UTC(),
	)

	if err != nil {
		// TODO: make sure this is the error is actually nonexistent commentHex
		return util.ErrorNoSuchComment
	}

	return nil
}

func commentList(commenter *data.UserCommenter, domain string, path string, isModerator bool) ([]*models.Comment, map[models.CommenterHexID]*models.Commenter, error) {
	// Prepare a query
	statement := "select commentHex, commenterHex, markdown, html, parentHex, score, state, deleted, creationDate " +
		"from comments " +
		"where comments.domain=$1 and comments.path=$2 and comments.deleted=false"
	params := []any{domain, path}

	// If the commenter is no moderator, show all unapproved comments
	if !isModerator {
		// Anonymous commenter: only include approved
		if commenter == nil {
			statement += " and comments.state='approved'"

		} else {
			// Authenticated commenter: also show their own unapproved comments
			statement += " and (comments.state='approved' or comments.commenterHex=$3)"
			params = append(params, commenter.HexID)
		}
	}
	statement += `;`

	// Fetch the comments
	rows, err := svc.DB.Query(statement, params...)
	if err != nil {
		logger.Errorf("cannot get comments: %v", err)
		return nil, nil, util.ErrorInternal
	}
	defer rows.Close()

	commenters := map[models.CommenterHexID]*models.Commenter{
		data.AnonymousCommenterHexID: {
			CommenterHex: data.AnonymousCommenterHexID,
			Email:        "undefined",
			Name:         "Anonymous",
			Link:         "undefined",
			Photo:        "undefined",
			Provider:     "undefined",
		},
	}

	var comments []*models.Comment
	for rows.Next() {
		comment := models.Comment{}
		if err = rows.Scan(
			&comment.CommentHex,
			&comment.CommenterHex,
			&comment.Markdown,
			&comment.HTML,
			&comment.ParentHex,
			&comment.Score,
			&comment.State,
			&comment.Deleted,
			&comment.CreationDate); err != nil {
			return nil, nil, util.ErrorInternal
		}

		// If it's an authenticated commenter, load their comment votes
		if commenter != nil {
			row := svc.DB.QueryRow(
				"select direction from votes where commentHex=$1 and commenterHex=$2;",
				comment.CommentHex,
				commenter.HexID)
			if err = row.Scan(&comment.Direction); err != nil {
				// TODO: is the only error here that there is no such entry?
				comment.Direction = 0
			}
		}

		// Do not include the original markdown for anonymous and other commenters, unless it's a moderator
		if commenter == nil || !isModerator && commenter.CommenterHexID() != comment.CommenterHex {
			comment.Markdown = ""
		}

		// Also, do not report comment state for non-moderators
		if !isModerator {
			comment.State = ""
		}

		// Append the comment to the list
		comments = append(comments, &comment)

		// Add the commenter to the map
		// TODO OMG this must be sloooooooow
		if _, ok := commenters[comment.CommenterHex]; !ok {
			if c, err := svc.TheUserService.FindCommenterByID(comment.CommenterHex); err != nil {
				return nil, nil, util.ErrorInternal
			} else {
				commenters[comment.CommenterHex] = c.ToCommenter()
			}
		}
	}

	// Succeeded
	return comments, commenters, nil
}
