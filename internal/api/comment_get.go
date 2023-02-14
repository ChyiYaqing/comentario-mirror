package api

import (
	"fmt"
	"gitlab.com/comentario/comentario/internal/api/models"
	"gitlab.com/comentario/comentario/internal/svc"
	"gitlab.com/comentario/comentario/internal/util"
)

var commentsRowColumns = `
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

func commentsRowScan(s sqlScanner, c *models.Comment) error {
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

func commentGetByCommentHex(commentHex string) (models.Comment, error) {
	if commentHex == "" {
		return models.Comment{}, util.ErrorMissingField
	}

	row := svc.DB.QueryRow(fmt.Sprintf(`select %s from comments where comments.commentHex = $1;`, commentsRowColumns), commentHex)

	var c models.Comment
	if err := commentsRowScan(row, &c); err != nil {
		// TODO: is this the only error?
		return c, util.ErrorNoSuchComment
	}

	return c, nil
}
