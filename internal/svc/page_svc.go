package svc

import (
	"database/sql"
	"fmt"
	"github.com/lib/pq"
	"gitlab.com/comentario/comentario/internal/api/models"
	"gitlab.com/comentario/comentario/internal/util"
	"strings"
)

// ThePageService is a global PageService implementation
var ThePageService PageService = &pageService{}

// PageService is a service interface for dealing with pages
type PageService interface {
	// CommentCountsByPath returns a map of comment counts by page path, for the specified domain and multiple paths
	CommentCountsByPath(domain string, paths []string) (map[string]int, error)
	// DeleteByDomain deletes all pages for the specified domain
	DeleteByDomain(domain string) error
	// FindByDomainPath finds and returns a pages for the specified domain and path combination. If no such page exists
	// in the database, return a new default Page model
	FindByDomainPath(domain, path string) (*models.Page, error)
	// UpdateTitleByDomainPath updates page title for the specified domain and path combination
	UpdateTitleByDomainPath(domain, path string) (string, error)
	// UpsertByDomainPath updates or inserts the page for the specified domain and path combination
	UpsertByDomainPath(domain, path string, isLocked bool, stickyCommentHex models.HexID) (*models.Page, error)
}

//----------------------------------------------------------------------------------------------------------------------

// pageService is a blueprint PageService implementation
type pageService struct{}

func (svc *pageService) CommentCountsByPath(domain string, paths []string) (map[string]int, error) {
	logger.Debugf("pageService.CommentCountsByPath(%s, ...)", domain)

	// Query paths/comment counts
	rows, err := db.Query("select path, commentcount from pages where domain=$1 and path=any($2);", domain, pq.Array(paths))
	if err != nil {
		logger.Errorf("pageService.CommentCountsByPath: Query() failed: %v", err)
		return nil, translateDBErrors(err)
	}
	defer rows.Close()

	// Fetch the paths and count, converting them into a map
	res := make(map[string]int)
	for rows.Next() {
		var p string
		var c int
		if err = rows.Scan(&p, &c); err != nil {
			logger.Errorf("pageService.CommentCountsByPath: rows.Scan() failed: %v", err)
			return nil, translateDBErrors(err)
		}
		res[p] = c
	}

	// Check that Next() didn't error
	if err := rows.Err(); err != nil {
		logger.Errorf("pageService.CommentCountsByPath: rows.Next() failed: %v", err)
		return nil, translateDBErrors(err)
	}

	// Succeeded
	return res, nil
}

func (svc *pageService) DeleteByDomain(domain string) error {
	logger.Debugf("pageService.DeleteByDomain(%s)", domain)

	// Delete records from the database
	if err := db.Exec("delete from pages where domain=$1;", domain); err != nil {
		logger.Errorf("pageService.DeleteByDomain: Exec() failed: %v", err)
		return translateDBErrors(err)
	}

	// Succeeded
	return nil
}

func (svc *pageService) FindByDomainPath(domain, path string) (*models.Page, error) {
	logger.Debugf("pageService.FindByDomainPath(%s, %s)", domain, path)

	// Query a page row
	row := db.QueryRow(
		"select domain, path, islocked, commentcount, stickycommenthex, title from pages where domain=$1 and path=$2;",
		domain,
		path)

	// Fetch the row
	var p models.Page
	if err := row.Scan(&p.Domain, &p.Path, &p.IsLocked, &p.CommentCount, &p.StickyCommentHex, &p.Title); err == sql.ErrNoRows {
		logger.Debug("pageService.FindByDomainPath: no page found, creating a new one")

		// No page in the database means there's no comment created yet for that page: make a default Page instance
		p.Domain = domain
		p.Path = path
		p.StickyCommentHex = "none" // TODO We really need to get rid of this

	} else if err != nil {
		// Any other database error
		logger.Errorf("pageService.FindByDomainPath: Scan() failed: %v", err)
		return nil, translateDBErrors(err)
	}

	// Succeeded
	return &p, nil
}

func (svc *pageService) UpdateTitleByDomainPath(domain, path string) (string, error) {
	logger.Debugf("pageService.UpdateTitleByDomainPath(%s, %s)", domain, path)

	// Try to fetch the title
	fullPath := fmt.Sprintf("%s/%s", domain, strings.TrimPrefix(path, "/"))
	title, err := util.HTMLTitleFromURL(fmt.Sprintf("http://%s", fullPath))

	// If fetching the title failed, just use domain/path combined as title
	if err != nil {
		title = fullPath
	}

	// Update the page in the database
	if err = db.Exec("update pages set title=$1 where domain=$2 and path=$3;", title, domain, path); err != nil {
		logger.Errorf("pageService.UpdateTitleByDomainPath: Exec() failed: %v", err)
		return "", translateDBErrors(err)
	}

	// Succeeded
	return title, nil
}

func (svc *pageService) UpsertByDomainPath(domain, path string, isLocked bool, stickyCommentHex models.HexID) (*models.Page, error) {
	logger.Debugf("pageService.UpsertByDomainPath(%s, %s, %v, %s)", domain, path, isLocked, stickyCommentHex)

	// Persist a new record, ignoring when it already exists
	if stickyCommentHex == "" {
		stickyCommentHex = "none" // TODO ditch this
	}
	page := models.Page{
		Domain:           domain,
		IsLocked:         isLocked,
		Path:             path,
		StickyCommentHex: stickyCommentHex,
	}
	err := db.Exec(
		"insert into pages(domain, path, islocked, stickycommenthex) values($1, $2, $3, $4) "+
			"on conflict (domain, path) do update set isLocked=$3, stickyCommentHex=$4;",
		page.Domain,
		page.Path,
		page.IsLocked,
		page.StickyCommentHex)
	if err != nil {
		logger.Errorf("pageService.UpsertByDomainPath: Exec() failed: %v", err)
		return nil, translateDBErrors(err)
	}

	// Succeeded
	return &page, nil
}
