package handlers

import (
	"database/sql"
	"github.com/go-openapi/runtime/middleware"
	"gitlab.com/comentario/comentario/internal/api/models"
	"gitlab.com/comentario/comentario/internal/api/restapi/operations"
	"gitlab.com/comentario/comentario/internal/svc"
	"gitlab.com/comentario/comentario/internal/util"
)

func PageUpdate(params operations.PageUpdateParams) middleware.Responder {
	commenter, err := commenterGetByCommenterToken(*params.Body.CommenterToken)
	if err != nil {
		return operations.NewPageUpdateOK().WithPayload(&models.APIResponseBase{Message: err.Error()})
	}

	isModerator, err := isDomainModerator(*params.Body.Domain, commenter.Email)
	if err != nil {
		return operations.NewPageUpdateOK().WithPayload(&models.APIResponseBase{Message: err.Error()})
	}

	if !isModerator {
		return operations.NewPageUpdateOK().WithPayload(&models.APIResponseBase{Message: util.ErrorNotModerator.Error()})
	}

	page := *params.Body.Attributes
	page.Domain = *params.Body.Domain
	page.Path = params.Body.Path

	if err = pageUpdate(&page); err != nil {
		return operations.NewPageUpdateOK().WithPayload(&models.APIResponseBase{Message: err.Error()})
	}

	// Succeeded
	return operations.NewPageUpdateOK().WithPayload(&models.APIResponseBase{Success: true})
}

func pageGet(domain string, path string) (*models.Page, error) {
	// Path can be empty
	if domain == "" {
		return nil, util.ErrorMissingField
	}

	row := svc.DB.QueryRow(
		"select isLocked, commentCount, stickyCommentHex, title from pages where domain=$1 and path=$2;",
		domain,
		path)

	p := &models.Page{Domain: domain, Path: path}
	if err := row.Scan(&p.IsLocked, &p.CommentCount, &p.StickyCommentHex, &p.Title); err != nil {
		if err == sql.ErrNoRows {
			// If there haven't been any comments, there won't be a record for this page. The sane thing to do is return
			// defaults
			// TODO: the defaults are hard-coded in two places: here and the schema
			p.IsLocked = false
			p.CommentCount = 0
			p.StickyCommentHex = "none"
			p.Title = ""
		} else {
			logger.Errorf("error scanning page: %v", err)
			return nil, util.ErrorInternal
		}
	}

	return p, nil
}

func pageNew(domain string, path string) error {
	// path can be empty
	if domain == "" {
		return util.ErrorMissingField
	}

	_, err := svc.DB.Exec("insert into pages(domain, path) values($1, $2) on conflict do nothing;", domain, path)
	if err != nil {
		logger.Errorf("error inserting new page: %v", err)
		return util.ErrorInternal
	}

	return nil
}

func pageTitleUpdate(domain string, path string) (string, error) {
	title, err := util.HTMLTitleGet("http://" + domain + path)
	if err != nil {
		// This could fail due to a variety of reasons that we can't control such as the user's URL 404 or something, so
		// let's not pollute the error log with messages. Just use a sane title. Maybe we'll have the ability to retry
		// later
		logger.Errorf("%v", err)
		title = domain
	}

	_, err = svc.DB.Exec("update pages set title = $3 where domain = $1 and path = $2;", domain, path, title)
	if err != nil {
		logger.Errorf("cannot update pages table with title: %v", err)
		return "", err
	}

	return title, nil
}

func pageUpdate(p *models.Page) error {
	if p.Domain == "" {
		return util.ErrorMissingField
	}

	// fields to not update:
	//   commentCount
	_, err := svc.DB.Exec(
		"insert into pages(domain, path, isLocked, stickyCommentHex) values($1, $2, $3, $4) "+
			"on conflict (domain, path) do update set isLocked = $3, stickyCommentHex = $4;",
		p.Domain,
		p.Path,
		p.IsLocked,
		p.StickyCommentHex)
	if err != nil {
		logger.Errorf("error setting page attributes: %v", err)
		return util.ErrorInternal
	}

	return nil
}
