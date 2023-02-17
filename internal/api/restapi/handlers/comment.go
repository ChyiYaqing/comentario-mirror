package handlers

import (
	"database/sql"
	"fmt"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/lib/pq"
	"gitlab.com/comentario/comentario/internal/api/models"
	"gitlab.com/comentario/comentario/internal/api/restapi/operations"
	"gitlab.com/comentario/comentario/internal/config"
	"gitlab.com/comentario/comentario/internal/svc"
	"gitlab.com/comentario/comentario/internal/util"
	"time"
)

const commentsRowColumns = `
	comments.commentHex,
	comments.commenterHex,
	comments.markdown,
	comments.html,
	comments.parentHex,
	comments.score,
	comments.state,
	comments.deleted,
	comments.creationDate
`

func CommentApprove(params operations.CommentApproveParams) middleware.Responder {
	c, err := commenterGetByCommenterToken(*params.Body.CommenterToken)
	if err != nil {
		return operations.NewCommentApproveOK().WithPayload(&models.APIResponseBase{Message: err.Error()})
	}

	domain, _, err := commentDomainPathGet(*params.Body.CommentHex)
	if err != nil {
		return operations.NewCommentApproveOK().WithPayload(&models.APIResponseBase{Message: err.Error()})
	}

	isModerator, err := isDomainModerator(domain, c.Email)
	if err != nil {
		return operations.NewCommentApproveOK().WithPayload(&models.APIResponseBase{Message: err.Error()})
	}

	if !isModerator {
		return operations.NewCommentApproveOK().WithPayload(&models.APIResponseBase{Message: util.ErrorNotModerator.Error()})
	}

	if err = commentApprove(*params.Body.CommentHex); err != nil {
		return operations.NewCommentApproveOK().WithPayload(&models.APIResponseBase{Message: err.Error()})
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
	commenter, err := commenterGetByCommenterToken(*params.Body.CommenterToken)
	if err != nil {
		return operations.NewCommentDeleteOK().WithPayload(&models.APIResponseBase{Message: err.Error()})
	}

	comment, err := commentGetByCommentHex(*params.Body.CommentHex)
	if err != nil {
		return operations.NewCommentDeleteOK().WithPayload(&models.APIResponseBase{Message: err.Error()})
	}

	domain, _, err := commentDomainPathGet(*params.Body.CommentHex)
	if err != nil {
		return operations.NewCommentDeleteOK().WithPayload(&models.APIResponseBase{Message: err.Error()})
	}

	isModerator, err := isDomainModerator(domain, commenter.Email)
	if err != nil {
		return operations.NewCommentDeleteOK().WithPayload(&models.APIResponseBase{Message: err.Error()})
	}

	if !isModerator && comment.CommenterHex != commenter.CommenterHex {
		return operations.NewCommentDeleteOK().WithPayload(&models.APIResponseBase{Message: util.ErrorNotModerator.Error()})
	}

	if err = commentDelete(*params.Body.CommentHex, commenter.CommenterHex); err != nil {
		return operations.NewCommentDeleteOK().WithPayload(&models.APIResponseBase{Message: err.Error()})
	}

	// Succeeded
	return operations.NewCommentDeleteOK().WithPayload(&models.APIResponseBase{Success: true})
}

func CommentEdit(params operations.CommentEditParams) middleware.Responder {
	commenter, err := commenterGetByCommenterToken(*params.Body.CommenterToken)
	if err != nil {
		return operations.NewCommentEditOK().WithPayload(&operations.CommentEditOKBody{Message: err.Error()})
	}

	comment, err := commentGetByCommentHex(*params.Body.CommentHex)
	if err != nil {
		return operations.NewCommentEditOK().WithPayload(&operations.CommentEditOKBody{Message: err.Error()})
	}

	if comment.CommenterHex != commenter.CommenterHex {
		return operations.NewCommentEditOK().WithPayload(&operations.CommentEditOKBody{Message: util.ErrorNotAuthorised.Error()})
	}

	html, err := commentEdit(*params.Body.CommentHex, *params.Body.Markdown)
	if err != nil {
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

	page, err := pageGet(domainName, params.Body.Path)
	if err != nil {
		return operations.NewCommentListOK().WithPayload(&operations.CommentListOKBody{Message: err.Error()})
	}

	var commenterHex models.HexID = "anonymous"
	isModerator := false
	modList := map[strfmt.Email]bool{}

	if *params.Body.CommenterToken != "anonymous" {
		c, err := commenterGetByCommenterToken(*params.Body.CommenterToken)
		if err != nil {
			if err != util.ErrorNoSuchToken {
				return operations.NewCommentListOK().WithPayload(&operations.CommentListOKBody{Message: err.Error()})
			}
			commenterHex = "anonymous"
		} else {
			commenterHex = c.CommenterHex
		}

		for _, mod := range domain.Moderators {
			modList[mod.Email] = true
			if mod.Email == c.Email {
				isModerator = true
			}
		}
	} else {
		for _, mod := range domain.Moderators {
			modList[mod.Email] = true
		}
	}

	domainViewRecord(domainName, commenterHex)

	comments, commenters, err := commentList(commenterHex, domainName, params.Body.Path, isModerator)
	if err != nil {
		return operations.NewCommentListOK().WithPayload(&operations.CommentListOKBody{Message: err.Error()})
	}

	_commenters := map[models.HexID]*models.Commenter{}
	for commenterHex, cr := range commenters {
		if _, ok := modList[cr.Email]; ok {
			cr.IsModerator = true
		}
		cr.Email = ""
		_commenters[commenterHex] = cr
	}

	return operations.NewCommentListOK().WithPayload(&operations.CommentListOKBody{
		Attributes: page,
		Commenters: _commenters,
		Comments:   comments,
		ConfiguredOauths: map[string]bool{
			"commento": domain.CommentoProvider,
			"google":   domain.GoogleProvider && config.OAuthGoogleConfig != nil,
			"github":   domain.GithubProvider && config.OAuthGithubConfig != nil,
			"gitlab":   domain.GitlabProvider && config.OAuthGitlabConfig != nil,
			"sso":      domain.SsoProvider,
		},
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

	if domain.RequireIdentification && *params.Body.CommenterToken == "anonymous" {
		return operations.NewCommentNewOK().WithPayload(&operations.CommentNewOKBody{Message: util.ErrorNotAuthorised.Error()})
	}

	commenterHex := models.HexID("anonymous")
	commenterEmail := strfmt.Email("")
	commenterName := "Anonymous"
	commenterLink := ""
	var isModerator bool
	if *params.Body.CommenterToken != "anonymous" {
		c, err := commenterGetByCommenterToken(*params.Body.CommenterToken)
		if err != nil {
			return operations.NewCommentNewOK().WithPayload(&operations.CommentNewOKBody{Message: err.Error()})
		}
		commenterHex = c.CommenterHex
		commenterEmail = c.Email
		commenterName = c.Name
		commenterLink = c.Link
		for _, mod := range domain.Moderators {
			if mod.Email == c.Email {
				isModerator = true
				break
			}
		}
	}

	var state models.CommentState
	if isModerator {
		state = models.CommentStateApproved
	} else if domain.RequireModeration || commenterHex == "anonymous" && domain.ModerateAllAnonymous {
		state = models.CommentStateUnapproved
	} else if domain.AutoSpamFilter && checkForSpam(*params.Body.Domain, util.UserIP(params.HTTPRequest), util.UserAgent(params.HTTPRequest), commenterName, string(commenterEmail), commenterLink, *params.Body.Markdown) {
		state = models.CommentStateFlagged
	} else {
		state = models.CommentStateApproved
	}

	commentHex, err := commentNew(commenterHex, domainName, params.Body.Path, *params.Body.ParentHex, *params.Body.Markdown, state, strfmt.DateTime(time.Now().UTC()))
	if err != nil {
		return operations.NewCommentNewOK().WithPayload(&operations.CommentNewOKBody{Message: err.Error()})
	}

	// TODO: reuse html in commentNew and do only one markdown to HTML conversion?
	html := util.MarkdownToHTML(*params.Body.Markdown)
	go emailNotificationNew(domain, params.Body.Path, commenterHex, commentHex, html, *params.Body.ParentHex, state)

	// Succeeded
	return operations.NewCommentNewOK().WithPayload(&operations.CommentNewOKBody{
		CommentHex: commentHex,
		HTML:       html,
		State:      state,
		Success:    true,
	})
}

func CommentVote(params operations.CommentVoteParams) middleware.Responder {
	if *params.Body.CommenterToken == "anonymous" {
		return operations.NewCommentVoteOK().WithPayload(&models.APIResponseBase{Message: util.ErrorUnauthorisedVote.Error()})
	}

	c, err := commenterGetByCommenterToken(*params.Body.CommenterToken)
	if err != nil {
		return operations.NewCommentVoteOK().WithPayload(&models.APIResponseBase{Message: err.Error()})
	}

	direction := 0
	if *params.Body.Direction > 0 {
		direction = 1
	} else if *params.Body.Direction < 0 {
		direction = -1
	}

	if err := commentVote(c.CommenterHex, *params.Body.CommentHex, direction); err != nil {
		return operations.NewCommentVoteOK().WithPayload(&models.APIResponseBase{Message: err.Error()})
	}

	// Succeeded
	return operations.NewCommentVoteOK().WithPayload(&models.APIResponseBase{Success: true})
}

func commentApprove(commentHex models.HexID) error {
	if commentHex == "" {
		return util.ErrorMissingField
	}

	_, err := svc.DB.Exec("update comments set state = 'approved' where commentHex = $1;", commentHex)
	if err != nil {
		logger.Errorf("cannot approve comment: %v", err)
		return util.ErrorInternal
	}

	return nil
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

	_, err := svc.DB.Exec(
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

func commentDomainPathGet(commentHex models.HexID) (string, string, error) {
	if commentHex == "" {
		return "", "", util.ErrorMissingField
	}

	row := svc.DB.QueryRow("select domain, path from comments where commentHex = $1;", commentHex)

	var domain string
	var path string
	var err error
	if err = row.Scan(&domain, &path); err != nil {
		return "", "", util.ErrorNoSuchDomain
	}

	return domain, path, nil
}

func commentEdit(commentHex models.HexID, markdown string) (string, error) {
	if commentHex == "" {
		return "", util.ErrorMissingField
	}

	html := util.MarkdownToHTML(markdown)
	if _, err := svc.DB.Exec("update comments set markdown = $2, html = $3 where commentHex=$1;", commentHex, markdown, html); err != nil {
		// TODO: make sure this is the error is actually nonexistent commentHex
		return "", util.ErrorNoSuchComment
	}

	return html, nil
}

func commentGetByCommentHex(commentHex models.HexID) (*models.Comment, error) {
	if commentHex == "" {
		return nil, util.ErrorMissingField
	}

	row := svc.DB.QueryRow(fmt.Sprintf(`select %s from comments where comments.commentHex = $1;`, commentsRowColumns), commentHex)

	var c models.Comment
	if err := commentsRowScan(row, &c); err != nil {
		// TODO: is this the only error?
		return nil, util.ErrorNoSuchComment
	}

	return &c, nil
}

func commentList(commenterHex models.HexID, domain string, path string, includeUnapproved bool) ([]*models.Comment, map[models.HexID]*models.Commenter, error) {
	// Path can be empty
	if commenterHex == "" || domain == "" {
		return nil, nil, util.ErrorMissingField
	}

	statement := "select commentHex, commenterHex, markdown, html, parentHex, score, state, deleted, creationDate " +
		"from comments " +
		"where comments.domain = $1 and comments.path = $2 and comments.deleted = false"

	if !includeUnapproved {
		if commenterHex == "anonymous" {
			statement += " and state = 'approved'"
		} else {
			statement += " and (state = 'approved' or commenterHex = $3)"
		}
	}

	statement += `;`

	var rows *sql.Rows
	var err error

	if !includeUnapproved && commenterHex != "anonymous" {
		rows, err = svc.DB.Query(statement, domain, path, commenterHex)
	} else {
		rows, err = svc.DB.Query(statement, domain, path)
	}

	if err != nil {
		logger.Errorf("cannot get comments: %v", err)
		return nil, nil, util.ErrorInternal
	}
	defer rows.Close()

	commenters := map[models.HexID]*models.Commenter{
		"anonymous": {
			CommenterHex: "anonymous",
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

		if commenterHex != "anonymous" {
			statement = `select direction from votes where commentHex=$1 and commenterHex=$2;`
			row := svc.DB.QueryRow(statement, comment.CommentHex, commenterHex)
			if err = row.Scan(&comment.Direction); err != nil {
				// TODO: is the only error here that there is no such entry?
				comment.Direction = 0
			}
		}

		if commenterHex != comment.CommenterHex {
			comment.Markdown = ""
		}

		if !includeUnapproved {
			comment.State = ""
		}

		comments = append(comments, &comment)

		if _, ok := commenters[comment.CommenterHex]; !ok {
			commenters[comment.CommenterHex], err = commenterGetByHex(comment.CommenterHex)
			if err != nil {
				logger.Errorf("cannot retrieve commenter: %v", err)
				return nil, nil, util.ErrorInternal
			}
		}
	}

	return comments, commenters, nil
}

// Take `creationDate` as a param because comment import (from Disqus, for example) will require a custom time
func commentNew(commenterHex models.HexID, domain string, path string, parentHex models.ParentHexID, markdown string, state models.CommentState, creationDate strfmt.DateTime) (models.HexID, error) {
	// path is allowed to be empty
	if commenterHex == "" || domain == "" || parentHex == "" || markdown == "" || state == "" {
		return "", util.ErrorMissingField
	}

	p, err := pageGet(domain, path)
	if err != nil {
		logger.Errorf("cannot get page attributes: %v", err)
		return "", util.ErrorInternal
	}

	if p.IsLocked {
		return "", util.ErrorThreadLocked
	}

	commentHex, err := util.RandomHex(32)
	if err != nil {
		return "", err
	}

	html := util.MarkdownToHTML(markdown)

	if err = pageNew(domain, path); err != nil {
		return "", err
	}

	_, err = svc.DB.Exec(
		"insert into comments(commentHex, domain, path, commenterHex, parentHex, markdown, html, creationDate, state) "+
			"values($1, $2, $3, $4, $5, $6, $7, $8, $9);",
		commentHex,
		domain,
		path,
		commenterHex,
		parentHex,
		markdown,
		html,
		creationDate,
		state)
	if err != nil {
		logger.Errorf("cannot insert comment: %v", err)
		return "", util.ErrorInternal
	}

	return models.HexID(commentHex), nil
}

func commentsRowScan(s util.Scanner, c *models.Comment) error {
	return s.Scan(
		&c.CommentHex,
		&c.CommenterHex,
		&c.Markdown,
		&c.HTML,
		&c.ParentHex,
		&c.Score,
		&c.State,
		&c.Deleted,
		&c.CreationDate,
	)
}

func commentVote(commenterHex models.HexID, commentHex models.HexID, direction int) error {
	if commentHex == "" || commenterHex == "" {
		return util.ErrorMissingField
	}

	row := svc.DB.QueryRow("select commenterHex from comments where commentHex = $1;", commentHex)

	var authorHex models.HexID
	if err := row.Scan(&authorHex); err != nil {
		logger.Errorf("error selecting authorHex for vote")
		return util.ErrorInternal
	}

	if authorHex == commenterHex {
		return util.ErrorSelfVote
	}

	_, err := svc.DB.Exec(
		"insert into votes(commentHex, commenterHex, direction, voteDate) values($1, $2, $3, $4) "+
			"on conflict (commentHex, commenterHex) do update set direction = $3;",
		commentHex,
		commenterHex,
		direction,
		time.Now().UTC())
	if err != nil {
		logger.Errorf("error inserting/updating votes: %v", err)
		return util.ErrorInternal
	}

	return nil
}
