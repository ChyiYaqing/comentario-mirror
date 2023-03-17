package svc

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/go-openapi/strfmt"
	"github.com/lunny/html2md"
	"gitlab.com/comentario/comentario/internal/api/models"
	"gitlab.com/comentario/comentario/internal/data"
	"gitlab.com/comentario/comentario/internal/util"
	"strings"
	"time"
)

// TheImportExportService is a global ImportExportService implementation
var TheImportExportService ImportExportService = &importExportService{}

// ImportExportService is a service interface for dealing with data import/export
type ImportExportService interface {
	// CreateExport exports the data for the specified domain, then adds and persists a new export record in the
	// database, returning its hex ID
	CreateExport(domain string) (models.HexID, error)
	// GetExportedData fetches an export with the given hex ID from the database, returning the related domain name, the
	// binary data, and the export timestamp
	GetExportedData(id models.HexID) (string, []byte, time.Time, error)
	// ImportCommento performs data import in the "commento" format, downloading the archive from the given URL. Returns
	// the number of imported comments
	ImportCommento(domain, dataURL string) (int64, error)
	// ImportDisqus performs data import from Disqus, downloading the archive from the given URL. Returns the number of
	// imported comments
	ImportDisqus(domain, dataURL string) (int64, error)
}

//----------------------------------------------------------------------------------------------------------------------

// importExportService is a blueprint ImportExportService implementation
type importExportService struct{}

type commentoExportV1 struct {
	Version    int                `json:"version"`
	Comments   []models.Comment   `json:"comments"`
	Commenters []models.Commenter `json:"commenters"`
}

type disqusThread struct {
	XMLName xml.Name `xml:"thread"`
	Id      string   `xml:"http://disqus.com/disqus-internals id,attr"`
	URL     string   `xml:"link"`
	Name    string   `xml:"name"`
}

type disqusAuthor struct {
	XMLName     xml.Name `xml:"author"`
	Name        string   `xml:"name"`
	IsAnonymous bool     `xml:"isAnonymous"`
	Username    string   `xml:"username"`
}

type disqusThreadId struct {
	XMLName xml.Name `xml:"thread"`
	Id      string   `xml:"http://disqus.com/disqus-internals id,attr"`
}

type disqusParentId struct {
	XMLName xml.Name `xml:"parent"`
	Id      string   `xml:"http://disqus.com/disqus-internals id,attr"`
}

type disqusPost struct {
	XMLName      xml.Name       `xml:"post"`
	Id           string         `xml:"http://disqus.com/disqus-internals id,attr"`
	ThreadId     disqusThreadId `xml:"thread"`
	ParentId     disqusParentId `xml:"parent"`
	Message      string         `xml:"message"`
	CreationDate time.Time      `xml:"createdAt"`
	IsDeleted    bool           `xml:"isDeleted"`
	IsSpam       bool           `xml:"isSpam"`
	Author       disqusAuthor   `xml:"author"`
}

type disqusXML struct {
	XMLName xml.Name       `xml:"disqus"`
	Threads []disqusThread `xml:"thread"`
	Posts   []disqusPost   `xml:"post"`
}

func (svc *importExportService) CreateExport(domain string) (models.HexID, error) {
	logger.Debugf("importExportService.CreateExport(%s)", domain)

	// Create an export data object
	exp := commentoExportV1{Version: 1}

	// Fetch comments
	var err error
	if exp.Comments, err = TheCommentService.ListByDomain(domain); err != nil {
		return "", err
	}

	// Fetch commenters
	if exp.Commenters, err = TheUserService.ListCommentersByDomain(domain); err != nil {
		return "", err
	}

	// Convert the data into JSON
	jsonData, err := json.Marshal(exp)
	if err != nil {
		logger.Errorf("importExportService.CreateExport: json.Marshal() failed: %v", err)
		return "", err
	}

	// Compress the JSON data with Gzip
	gzippedData, err := util.GzipStatic(jsonData)
	if err != nil {
		logger.Errorf("importExportService.CreateExport: GzipStatic() failed: %v", err)
		return "", err
	}

	// Generate a new random hex ID for the export
	id, err := data.RandomHexID()
	if err != nil {
		logger.Errorf("importExportService.CreateExport: RandomHexID() failed: %v", err)
		return "", err
	}

	// Insert a database record
	err = db.Exec(
		"insert into exports(exporthex, bindata, domain, creationdate) values($1, $2, $3, $4);",
		id, gzippedData, domain, time.Now().UTC())
	if err != nil {
		logger.Errorf("importExportService.CreateExport: Exec() failed: %v", err)
		return "", translateDBErrors(err)
	}

	// Succeeded
	return id, nil
}

func (svc *importExportService) GetExportedData(id models.HexID) (string, []byte, time.Time, error) {
	logger.Debugf("importExportService.GetExportedData(%s)", id)

	// Fetch the data record
	row := db.QueryRow("select domain, bindata, creationdate from exports where exporthex=$1;", id)
	var domain string
	var binData []byte
	var created time.Time
	if err := row.Scan(&domain, &binData, &created); err != nil {
		logger.Errorf("importExportService.GetExportedData: QueryRow failed(): %v", err)
		return "", nil, time.Time{}, translateDBErrors(err)
	}

	// Succeeded
	return domain, binData, created, nil
}

func (svc *importExportService) ImportCommento(domain, dataURL string) (int64, error) {
	logger.Debugf("importExportService.ImportCommento(%s, %s)", domain, dataURL)

	// Validate the URL
	if _, err := util.ParseAbsoluteURL(dataURL); err != nil {
		return 0, err
	}

	// Fetch and decompress the export tarball
	d, err := util.DownloadGzip(dataURL)
	if err != nil {
		logger.Errorf("importExportService.ImportCommento: DownloadGzip() failed: %v", err)
		return 0, err
	}

	// Unmarshal the data
	var exp commentoExportV1
	if err := json.Unmarshal(d, &exp); err != nil {
		logger.Errorf("importExportService.ImportCommento: json.Unmarshal() failed: %v", err)
		return 0, err
	}

	// Verify the export format version
	if exp.Version != 1 {
		logger.Errorf("importExportService.ImportCommento: invalid export version (got %d, want 1)", exp.Version)
		return 0, util.ErrorBadCommentoExportVersion
	}

	// Check if imported commentedHex or email exists, creating a map of commenterHex (old hex, new hex)
	commenterHex := map[models.CommenterHexID]models.CommenterHexID{data.AnonymousCommenterHexID: data.AnonymousCommenterHexID}
	for _, commenter := range exp.Commenters {
		// Try to find an existing commenter with the same email
		if c, err := TheUserService.FindCommenterByIdPEmail("", string(commenter.Email), false); err == nil {
			// Commenter already exists. Add its hex ID to the map and proceed to the next record
			commenterHex[commenter.CommenterHex] = c.CommenterHexID()
			continue
		} else if err != ErrNotFound {
			// Any other error than "not found"
			return 0, err
		}

		// Generate a random password string
		randomPassword, err := data.RandomHexID()
		if err != nil {
			logger.Errorf("importExportService.ImportCommento: RandomHexID() failed: %v", err)
			return 0, util.ErrorInternal
		}

		// Persist a new commenter instance
		if c, err := TheUserService.CreateCommenter(string(commenter.Email), commenter.Name, commenter.Link, commenter.Photo, "", string(randomPassword)); err != nil {
			return 0, err
		} else {
			// Save the new commenter's hex ID in the map
			commenterHex[commenter.CommenterHex] = c.CommenterHexID()
		}
	}

	// Create a map of (parent hex, comments)
	comments := map[models.ParentHexID][]models.Comment{}
	for _, comment := range exp.Comments {
		comments[comment.ParentHex] = append(comments[comment.ParentHex], comment)
	}

	// Import comments, creating a map of comment hex (old hex, new hex)
	commentHex := map[models.ParentHexID]models.ParentHexID{data.RootParentHexID: data.RootParentHexID}
	count := int64(0)
	keys := []models.ParentHexID{data.RootParentHexID}
	for i := 0; i < len(keys); i++ {
		for _, comment := range comments[keys[i]] {
			// Find the comment's author
			cHex, ok := commenterHex[comment.CommenterHex]
			if !ok {
				logger.Errorf("importExportService.ImportCommento: failed to find mapped commenter (hex=%v)", comment.CommenterHex)
				return count, util.ErrorInternal
			}

			// Find the parent comment
			parentHex, ok := commentHex[comment.ParentHex]
			if !ok {
				logger.Errorf("importExportService.ImportCommento: failed to find parent comment (hex=%v)", comment.ParentHex)
				return count, util.ErrorInternal
			}

			// Add a new comment record
			newComment, err := TheCommentService.Create(cHex, domain, comment.URL, comment.Markdown, parentHex, comment.State, comment.CreationDate)
			if err != nil {
				return count, err
			}

			// Store the comment's hex ID in the map
			commentHex[models.ParentHexID(comment.CommentHex)] = models.ParentHexID(newComment.CommentHex)
			keys = append(keys, models.ParentHexID(comment.CommentHex))

			// Import record counter
			count++
		}
	}

	// Succeeded
	return count, nil
}

func (svc *importExportService) ImportDisqus(domain, dataURL string) (int64, error) {
	logger.Debugf("importExportService.ImportDisqus(%s, %s)", domain, dataURL)

	// Validate the URL and check it's on *.discus.com
	if u, err := util.ParseAbsoluteURL(dataURL); err != nil {
		return 0, err
	} else if u.Host != "disqus.com" && !strings.HasSuffix(u.Host, ".disqus.com") {
		return 0, util.ErrorNoDisqusURL
	}

	// Fetch and decompress the export tarball
	d, err := util.DownloadGzip(dataURL)
	if err != nil {
		logger.Errorf("importExportService.ImportDisqus: DownloadGzip() failed: %v", err)
		return 0, err
	}

	// Unmarshal the XML data
	exp := disqusXML{}
	err = xml.Unmarshal(d, &exp)
	if err != nil {
		logger.Errorf("importExportService.ImportDisqus: xml.Unmarshal() failed: %v", err)
		return 0, err
	}

	// Map Disqus thread IDs to threads
	threads := make(map[string]disqusThread)
	for _, thread := range exp.Threads {
		threads[thread.Id] = thread
	}

	// Map Disqus emails to commenterHex (if not available, create a new one with a random password that can be reset
	// later)
	commenterHex := map[string]models.CommenterHexID{}
	for _, post := range exp.Posts {
		if post.IsDeleted || post.IsSpam {
			continue
		}

		// Skip authors whose email has already been processed
		email := fmt.Sprintf("%s@disqus.com", post.Author.Username)
		if _, ok := commenterHex[email]; ok {
			continue
		}

		// Try to find an existing commenter with this email
		if c, err := TheUserService.FindCommenterByIdPEmail("", email, false); err == nil {
			// Commenter already exists. Add its hex ID to the map and proceed to the next record
			commenterHex[email] = c.CommenterHexID()
			continue
		} else if err != ErrNotFound {
			// Any other error than "not found"
			return 0, err
		}

		// Generate a random password string
		randomPassword, err := data.RandomHexID()
		if err != nil {
			logger.Errorf("importExportService.ImportDisqus: RandomHexID() failed: %v", err)
			return 0, util.ErrorInternal
		}

		// Persist a new commenter instance
		if c, err := TheUserService.CreateCommenter(email, post.Author.Name, "", "", "", string(randomPassword)); err != nil {
			return 0, err
		} else {
			// Save the new commenter's hex ID in the map
			commenterHex[email] = c.CommenterHexID()
		}
	}

	// For each Disqus post, create a Comentario comment
	count := int64(0)
	disqusIdMap := make(map[string]models.HexID)
	for _, post := range exp.Posts {
		// Skip over deleted and spam posts
		if post.IsDeleted || post.IsSpam {
			continue
		}

		// Find the commenter hex ID by their email
		cHex := data.AnonymousCommenterHexID
		if !post.Author.IsAnonymous {
			cHex = commenterHex[fmt.Sprintf("%s@disqus.com", post.Author.Username)]
		}

		parentHex := data.RootParentHexID
		if val, ok := disqusIdMap[post.ParentId.Id]; ok {
			parentHex = models.ParentHexID(val)
		}

		// Extract the path from thread URL
		var path string
		if u, err := util.ParseAbsoluteURL(threads[post.ThreadId.Id].URL); err != nil {
			return count, err
		} else {
			path = u.Path
		}

		// Create a new post record
		// TODO restrict the list of tags to just the basics: <a>, <b>, <i>, <code>. Especially remove <img> (convert it to <a>)
		comment, err := TheCommentService.Create(
			cHex,
			domain,
			path,
			html2md.Convert(post.Message),
			parentHex,
			models.CommentStateApproved,
			strfmt.DateTime(post.CreationDate))
		if err != nil {
			return count, err
		}

		// Add the comment's hex ID to the ID map
		disqusIdMap[post.Id] = comment.CommentHex

		// Import record counter
		count++
	}

	// Succeeded
	return count, nil
}
