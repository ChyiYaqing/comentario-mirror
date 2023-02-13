package api

import (
	"database/sql"
	"gitlab.com/comentario/comentario/internal/svc"
	"gitlab.com/comentario/comentario/internal/util"
)

func pageGet(domain string, path string) (page, error) {
	// path can be empty
	if domain == "" {
		return page{}, util.ErrorMissingField
	}

	statement := `
		select isLocked, commentCount, stickyCommentHex, title
		from pages
		where domain=$1 and path=$2;
	`
	row := svc.DB.QueryRow(statement, domain, path)

	p := page{Domain: domain, Path: path}
	if err := row.Scan(&p.IsLocked, &p.CommentCount, &p.StickyCommentHex, &p.Title); err != nil {
		if err == sql.ErrNoRows {
			// If there haven't been any comments, there won't be a record for this
			// page. The sane thing to do is return defaults.
			// TODO: the defaults are hard-coded in two places: here and the schema
			p.IsLocked = false
			p.CommentCount = 0
			p.StickyCommentHex = "none"
			p.Title = ""
		} else {
			logger.Errorf("error scanning page: %v", err)
			return page{}, util.ErrorInternal
		}
	}

	return p, nil
}
