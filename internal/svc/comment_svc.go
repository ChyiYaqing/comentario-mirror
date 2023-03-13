package svc

import (
	"gitlab.com/comentario/comentario/internal/api/models"
)

// TheCommentService is a global CommentService implementation
var TheCommentService CommentService = &commentService{}

// CommentService is a service interface for dealing with comments
type CommentService interface {
	// FindByHexID finds and returns a comment with the given hex ID
	FindByHexID(commentHex models.HexID) (*models.Comment, error)
	// UpdateText updates the markdown and the HTML of a comment with the given hex ID in the database
	UpdateText(commentHex models.HexID, markdown, html string) error
}

//----------------------------------------------------------------------------------------------------------------------

// commentService is a blueprint CommentService implementation
type commentService struct{}

func (svc *commentService) FindByHexID(commentHex models.HexID) (*models.Comment, error) {
	// Validate the passed ID
	if err := validateHexID(commentHex); err != nil {
		return nil, err
	}

	// Query the database
	row := db.QueryRow(
		"select commentHex, commenterHex, markdown, html, parentHex, score, state, deleted, creationDate "+
			"from comments "+
			"where commentHex=$1;",
		commentHex)

	// Fetch the comment
	var c models.Comment
	err := row.Scan(&c.CommentHex, &c.CommenterHex, &c.Markdown, &c.HTML, &c.ParentHex, &c.Score, &c.State, &c.Deleted, &c.CreationDate)
	if err != nil {
		return nil, translateErrors(err)
	}

	// Succeeded
	return &c, nil
}

func (svc *commentService) UpdateText(commentHex models.HexID, markdown, html string) error {
	// Validate the passed ID
	if err := validateHexID(commentHex); err != nil {
		return err
	}

	// Update the row in the database
	if _, err := db.Exec("update comments set markdown=$1, html=$2 where commentHex=$3;", markdown, html, commentHex); err != nil {
		return translateErrors(err)
	}

	// Succeeded
	return nil
}
