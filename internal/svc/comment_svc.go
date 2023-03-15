package svc

import (
	"github.com/go-openapi/strfmt"
	"gitlab.com/comentario/comentario/internal/api/models"
	"gitlab.com/comentario/comentario/internal/data"
	"gitlab.com/comentario/comentario/internal/util"
	"time"
)

// TheCommentService is a global CommentService implementation
var TheCommentService CommentService = &commentService{}

// CommentService is a service interface for dealing with comments
type CommentService interface {
	// Approve sets the status of a comment with the given hex ID to 'approved'
	Approve(commentHex models.HexID) error
	// Create creates, persists, and returns a new comment
	Create(commenterHex models.CommenterHexID, domain, path, markdown string, parentHex models.ParentHexID, state models.CommentState, creationDate strfmt.DateTime) (*models.Comment, error)
	// DeleteByDomain deletes all comments for the specified domain
	DeleteByDomain(domain string) error
	// FindByHexID finds and returns a comment with the given hex ID
	FindByHexID(commentHex models.HexID) (*models.Comment, error)
	// ListByDomainPath returns a list of comments and related commenters for the given domain and path combination.
	// commenterHex should be "anonymous" for an unauthenticated user
	ListByDomainPath(commenterHex models.CommenterHexID, domain, path string, isModerator bool) ([]*models.Comment, map[models.CommenterHexID]*models.Commenter, error)
	// MarkDeleted mark a comment with the given hex ID deleted in the database
	MarkDeleted(commentHex models.HexID, deleterHex models.CommenterHexID) error
	// UpdateText updates the markdown and the HTML of a comment with the given hex ID in the database
	UpdateText(commentHex models.HexID, markdown, html string) error
}

//----------------------------------------------------------------------------------------------------------------------

// commentService is a blueprint CommentService implementation
type commentService struct{}

func (svc *commentService) Approve(commentHex models.HexID) error {
	logger.Debugf("commentService.Approve(%s)", commentHex)

	// Update the record in the database
	if err := db.Exec("update comments set state=$1 where commentHex=$2;", models.CommentStateApproved, commentHex); err != nil {
		logger.Errorf("commentService.Approve: Exec() failed: %v", err)
		return translateErrors(err)
	}

	// Succeeded
	return nil
}

func (svc *commentService) Create(commenterHex models.CommenterHexID, domain, path, markdown string, parentHex models.ParentHexID, state models.CommentState, creationDate strfmt.DateTime) (*models.Comment, error) {
	logger.Debugf("commentService.Create(%s, %s, %s, ...)", commenterHex, domain, path)

	// Fetch the related page
	page, err := ThePageService.FindByDomainPath(domain, path)
	if err != nil {
		return nil, err
	}

	// Make sure the page isn't locked
	if page.IsLocked {
		return nil, ErrPageLocked
	}

	// Generate a new comment hex ID
	commentHex, err := util.RandomHex(32)
	if err != nil {
		return nil, err
	}

	// Convert the markdown into HTML
	html := util.MarkdownToHTML(markdown)

	// Persist a new page record (if necessary)
	if _, err = ThePageService.UpsertByDomainPath(domain, path, false, ""); err != nil {
		return nil, err
	}

	// Persist a new comment record
	c := models.Comment{
		CommentHex:   models.HexID(commentHex),
		CommenterHex: commenterHex,
		CreationDate: creationDate,
		Domain:       domain,
		HTML:         html,
		Markdown:     markdown,
		ParentHex:    parentHex,
		State:        state,
		URL:          path,
	}
	err = db.Exec(
		"insert into comments(commentHex, domain, path, commenterHex, parentHex, markdown, html, creationDate, state) "+
			"values($1, $2, $3, $4, $5, $6, $7, $8, $9);",
		c.CommentHex, c.Domain, c.URL, c.CommenterHex, c.ParentHex, c.Markdown, c.HTML, c.CreationDate, c.State)
	if err != nil {
		logger.Errorf("commentService.Create: Exec() failed: %v", err)
		return nil, translateErrors(err)
	}

	return &c, nil
}

func (svc *commentService) DeleteByDomain(domain string) error {
	logger.Debugf("commentService.DeleteByDomain(%s)", domain)

	// Delete records from the database
	if err := db.Exec("delete from comments where domain=$1;", domain); err != nil {
		logger.Errorf("commentService.DeleteByDomain: Exec() failed: %v", err)
		return translateErrors(err)
	}

	// Succeeded
	return nil
}

func (svc *commentService) FindByHexID(commentHex models.HexID) (*models.Comment, error) {
	logger.Debugf("commentService.FindByHexID(%s)", commentHex)

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

func (svc *commentService) ListByDomainPath(commenterHex models.CommenterHexID, domain, path string, isModerator bool) ([]*models.Comment, map[models.CommenterHexID]*models.Commenter, error) {
	logger.Debugf("commentService.ListByDomainPath(%s, %s, %s, %v)", commenterHex, domain, path, isModerator)

	// Prepare a query
	statement :=
		"select " +
			"c.commenthex, c.commenterhex, c.markdown, c.html, c.parenthex, c.score, c.state, c.deleted, c.creationdate, " +
			"coalesce(v.direction, 0), " +
			"coalesce(r.commenterHex, ''), " +
			"coalesce(r.email, ''), " +
			"coalesce(r.name, ''), " +
			"coalesce(r.link, ''), " +
			"coalesce(r.photo, ''), " +
			"coalesce(r.provider, ''), " +
			"coalesce(r.joinDate, CURRENT_TIMESTAMP) " +
			"from comments c " +
			"left join votes v on v.commenthex=c.commenthex and v.commenterhex=$1 " +
			"left join commenters r on r.commenterhex=c.commenterhex " +
			"where c.domain=$2 and c.path=$3 and c.deleted=false"
	params := []any{commenterHex, domain, path}

	// Anonymous commenter: only include approved
	if commenterHex == data.AnonymousCommenterHexID {
		statement += " and c.state=$4"
		params = append(params, models.CommentStateApproved)

	} else if !isModerator {
		// Authenticated, non-moderator commenter: show only approved and all own comments
		statement += " and (c.state=$4 or c.commenterHex=$1)"
		params = append(params, models.CommentStateApproved)
	}
	statement += ";"

	// Fetch the comments
	rs, err := db.Query(statement, params...)
	if err != nil {
		logger.Errorf("commentService.ListByDomainPath: Query() failed: %v", err)
		return nil, nil, util.ErrorInternal
	}
	defer rs.Close()

	// Prepare commenter map: begin with only the "anonymous" one
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

	// Iterate result rows
	var comments []*models.Comment
	for rs.Next() {
		// Fetch the comment and the related commenter
		comment := models.Comment{}
		uc := data.UserCommenter{}
		err := rs.Scan(
			&comment.CommentHex,
			&comment.CommenterHex,
			&comment.Markdown,
			&comment.HTML,
			&comment.ParentHex,
			&comment.Score,
			&comment.State,
			&comment.Deleted,
			&comment.CreationDate,
			&comment.Direction,
			&uc.HexID,
			&uc.Email,
			&uc.Name,
			&uc.WebsiteURL,
			&uc.PhotoURL,
			&uc.Provider,
			&uc.Created)
		if err != nil {
			logger.Errorf("commentService.ListByDomainPath: Scan() failed: %v", err)
			return nil, nil, translateErrors(err)
		}

		// Do not include the original markdown for anonymous and other commenters, unless it's a moderator
		if commenterHex == data.AnonymousCommenterHexID || !isModerator && commenterHex != comment.CommenterHex {
			comment.Markdown = ""
		}

		// Also, do not report comment state for non-moderators
		if !isModerator {
			comment.State = ""
		}

		// Append the comment to the list
		comments = append(comments, &comment)

		// Add the commenter, if any, to the map
		if uc.HexID != "" {
			if _, ok := commenters[comment.CommenterHex]; !ok {
				commenters[comment.CommenterHex] = uc.ToCommenter()
			}
		}
	}

	// Check if Next() didn't error
	if err := rs.Err(); err != nil {
		return nil, nil, err
	}

	// Succeeded
	return comments, commenters, nil
}

func (svc *commentService) MarkDeleted(commentHex models.HexID, deleterHex models.CommenterHexID) error {
	logger.Debugf("commentService.MarkDeleted(%s, %s)", commentHex, deleterHex)

	// Update the record in the database
	err := db.Exec(
		"update comments "+
			"set deleted=true, markdown='[deleted]', html='[deleted]', deleterhex=$1, deletiondate=$2 "+
			"where commenthex=$3;",
		deleterHex,
		time.Now().UTC(),
		commentHex)
	if err != nil {
		logger.Errorf("commentService.MarkDeleted: Exec() failed: %v", err)
		return translateErrors(err)
	}

	// Succeeded
	return nil
}

func (svc *commentService) UpdateText(commentHex models.HexID, markdown, html string) error {
	logger.Debugf("commentService.UpdateText(%s, ...)", commentHex)

	// Update the row in the database
	if err := db.Exec("update comments set markdown=$1, html=$2 where commentHex=$3;", markdown, html, commentHex); err != nil {
		logger.Errorf("commentService.UpdateText: Exec() failed: %v", err)
		return translateErrors(err)
	}

	// Succeeded
	return nil
}
