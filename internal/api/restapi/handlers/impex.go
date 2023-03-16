package handlers

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/lunny/html2md"
	"gitlab.com/comentario/comentario/internal/api/models"
	"gitlab.com/comentario/comentario/internal/api/restapi/operations"
	"gitlab.com/comentario/comentario/internal/config"
	"gitlab.com/comentario/comentario/internal/data"
	"gitlab.com/comentario/comentario/internal/svc"
	"gitlab.com/comentario/comentario/internal/util"
	"io"
	"net/http"
	"regexp"
	"time"
)

type commentoExportV1 struct {
	Version    int                `json:"version"`
	Comments   []models.Comment   `json:"comments"`
	Commenters []models.Commenter `json:"commenters"`
}

func DomainExportBegin(params operations.DomainExportBeginParams) middleware.Responder {
	if !config.SMTPConfigured {
		return operations.NewDomainExportBeginOK().WithPayload(&models.APIResponseBase{Message: util.ErrorSmtpNotConfigured.Error()})
	}

	user, err := svc.TheUserService.FindOwnerByToken(*params.Body.OwnerToken)
	if err != nil {
		return operations.NewDomainExportBeginOK().WithPayload(&models.APIResponseBase{Message: err.Error()})
	}

	isOwner, err := domainOwnershipVerify(user.HexID, *params.Body.Domain)
	if err != nil {
		return operations.NewDomainExportBeginOK().WithPayload(&models.APIResponseBase{Message: err.Error()})
	}

	if !isOwner {
		return operations.NewDomainExportBeginOK().WithPayload(&models.APIResponseBase{Message: util.ErrorNotAuthorised.Error()})
	}

	go domainExportBegin(strfmt.Email(user.Email), *params.Body.Domain)

	// Succeeded
	return operations.NewDomainExportBeginOK().WithPayload(&models.APIResponseBase{Success: true})
}

func DomainExportDownload(params operations.DomainExportDownloadParams) middleware.Responder {
	row := svc.DB.QueryRow("select domain, binData, creationDate from exports where exportHex = $1;", params.ExportHex)
	var domain string
	var binData []byte
	var creationDate time.Time
	if err := row.Scan(&domain, &binData, &creationDate); err != nil {
		return operations.NewGenericNotFound().WithPayload(&operations.GenericNotFoundBody{Details: "Wrong exportHex value"})
	}

	// Succeeded
	return operations.NewDomainExportDownloadOK().
		WithContentDisposition(fmt.Sprintf(`inline; filename="%s-%v.json.gz"`, domain, creationDate.Unix())).
		WithPayload(io.NopCloser(bytes.NewReader(binData)))
}

func DomainImportCommento(params operations.DomainImportCommentoParams) middleware.Responder {
	user, err := svc.TheUserService.FindOwnerByToken(*params.Body.OwnerToken)
	if err != nil {
		return operations.NewDomainImportCommentoOK().WithPayload(&operations.DomainImportCommentoOKBody{Message: err.Error()})
	}

	domainName := *params.Body.Domain
	isOwner, err := domainOwnershipVerify(user.HexID, domainName)
	if err != nil {
		return operations.NewDomainImportCommentoOK().WithPayload(&operations.DomainImportCommentoOKBody{Message: err.Error()})
	}
	if !isOwner {
		return operations.NewDomainImportCommentoOK().WithPayload(&operations.DomainImportCommentoOKBody{Message: util.ErrorNotAuthorised.Error()})
	}

	numImported, err := domainImportCommento(domainName, *params.Body.URL)
	if err != nil {
		return operations.NewDomainImportCommentoOK().WithPayload(&operations.DomainImportCommentoOKBody{Message: err.Error()})
	}

	// Succeeded
	return operations.NewDomainImportCommentoOK().WithPayload(&operations.DomainImportCommentoOKBody{
		NumImported: int64(numImported),
		Success:     true,
	})
}

func DomainImportDisqus(params operations.DomainImportDisqusParams) middleware.Responder {
	user, err := svc.TheUserService.FindOwnerByToken(*params.Body.OwnerToken)
	if err != nil {
		return operations.NewDomainImportDisqusOK().WithPayload(&operations.DomainImportDisqusOKBody{Message: err.Error()})
	}

	domainName := *params.Body.Domain
	isOwner, err := domainOwnershipVerify(user.HexID, domainName)
	if err != nil {
		return operations.NewDomainImportDisqusOK().WithPayload(&operations.DomainImportDisqusOKBody{Message: err.Error()})
	}
	if !isOwner {
		return operations.NewDomainImportDisqusOK().WithPayload(&operations.DomainImportDisqusOKBody{Message: util.ErrorNotAuthorised.Error()})
	}

	numImported, err := domainImportDisqus(domainName, *params.Body.URL)
	if err != nil {
		return operations.NewDomainImportDisqusOK().WithPayload(&operations.DomainImportDisqusOKBody{Message: err.Error()})
	}

	// Succeeded
	return operations.NewDomainImportDisqusOK().WithPayload(&operations.DomainImportDisqusOKBody{
		NumImported: int64(numImported),
		Success:     true,
	})
}

// domainExportBeginError notifies the user by email about an export error
func domainExportBeginError(email strfmt.Email, domain string, err error) {
	_ = svc.TheMailService.SendFromTemplate(
		"",
		string(email),
		"Comentario Data Export Errored",
		"domain-export-error.gohtml",
		map[string]any{"Domain": domain, "Error": err.Error()})
}

func domainExportBegin(email strfmt.Email, domain string) {
	e := commentoExportV1{Version: 1, Comments: []models.Comment{}, Commenters: []models.Commenter{}}
	rows1, err := svc.DB.Query(
		"select commentHex, domain, path, commenterHex, markdown, parentHex, score, state, creationDate from comments where domain = $1;",
		domain)
	if err != nil {
		logger.Errorf("cannot select comments while exporting %s: %v", domain, err)
		domainExportBeginError(email, domain, util.ErrorInternal)
		return
	}
	defer rows1.Close()

	for rows1.Next() {
		c := models.Comment{}
		if err = rows1.Scan(&c.CommentHex, &c.Domain, &c.URL, &c.CommenterHex, &c.Markdown, &c.ParentHex, &c.Score, &c.State, &c.CreationDate); err != nil {
			logger.Errorf("cannot scan comment while exporting %s: %v", domain, err)
			domainExportBeginError(email, domain, util.ErrorInternal)
			return
		}

		e.Comments = append(e.Comments, c)
	}

	rows2, err := svc.DB.Query(
		"select commenters.commenterHex, commenters.email, commenters.name, commenters.link, commenters.photo, commenters.provider, commenters.joinDate "+
			"from commenters, comments "+
			"where comments.domain = $1 and commenters.commenterHex = comments.commenterHex;",
		domain)
	if err != nil {
		logger.Errorf("cannot select commenters while exporting %s: %v", domain, err)
		domainExportBeginError(email, domain, util.ErrorInternal)
		return
	}
	defer rows2.Close()

	for rows2.Next() {
		c := models.Commenter{}
		if err := rows2.Scan(&c.CommenterHex, &c.Email, &c.Name, &c.Link, &c.Photo, &c.Provider, &c.JoinDate); err != nil {
			logger.Errorf("cannot scan commenter while exporting %s: %v", domain, err)
			domainExportBeginError(email, domain, util.ErrorInternal)
			return
		}

		e.Commenters = append(e.Commenters, c)
	}

	je, err := json.Marshal(e)
	if err != nil {
		logger.Errorf("cannot marshall JSON while exporting %s: %v", domain, err)
		domainExportBeginError(email, domain, util.ErrorInternal)
		return
	}

	gje, err := util.GzipStatic(je)
	if err != nil {
		logger.Errorf("cannot gzip JSON while exporting %s: %v", domain, err)
		domainExportBeginError(email, domain, util.ErrorInternal)
		return
	}

	exportHex, err := data.RandomHexID()
	if err != nil {
		logger.Errorf("cannot generate exportHex while exporting %s: %v", domain, err)
		domainExportBeginError(email, domain, util.ErrorInternal)
		return
	}

	err = svc.DB.Exec(
		"insert into exports(exportHex, binData, domain, creationDate) values($1, $2, $3, $4);",
		exportHex,
		gje,
		domain,
		time.Now().UTC())
	if err != nil {
		logger.Errorf("error inserting expiry binary data while exporting %s: %v", domain, err)
		domainExportBeginError(email, domain, util.ErrorInternal)
		return
	}

	// Notify the user by email, ignoring any error
	_ = svc.TheMailService.SendFromTemplate(
		"",
		string(email),
		"Comentario Data Export",
		"domain-export.gohtml",
		map[string]any{
			"Domain": domain,
			"URL":    config.URLForAPI("domain/export/download", map[string]string{"exportHex": string(exportHex)}),
		})
}

func domainImportCommento(domain string, url string) (int, error) {
	if domain == "" || url == "" {
		return 0, util.ErrorMissingField
	}

	resp, err := http.Get(url)
	if err != nil {
		logger.Errorf("cannot get url: %v", err)
		return 0, util.ErrorCannotDownloadCommento
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Errorf("cannot read body: %v", err)
		return 0, util.ErrorCannotDownloadCommento
	}

	zr, err := gzip.NewReader(bytes.NewBuffer(body))
	if err != nil {
		logger.Errorf("cannot create gzip reader: %v", err)
		return 0, util.ErrorInternal
	}

	contents, err := io.ReadAll(zr)
	if err != nil {
		logger.Errorf("cannot read gzip contents uncompressed: %v", err)
		return 0, util.ErrorInternal
	}

	var inputData commentoExportV1
	if err := json.Unmarshal(contents, &inputData); err != nil {
		logger.Errorf("cannot unmarshal JSON at %s: %v", url, err)
		return 0, util.ErrorInternal
	}

	if inputData.Version != 1 {
		logger.Errorf("invalid data version (got %d, want 1): %v", inputData.Version, err)
		return 0, util.ErrorUnsupportedCommentoImportVersion
	}

	// Check if imported commentedHex or email exists, creating a map of commenterHex (old hex, new hex)
	commenterHex := map[models.CommenterHexID]models.CommenterHexID{data.AnonymousCommenterHexID: data.AnonymousCommenterHexID}
	for _, commenter := range inputData.Commenters {
		// Try to find an existing commenter with the same email
		if c, err := svc.TheUserService.FindCommenterByIdPEmail("", string(commenter.Email), false); err == nil {
			// Commenter already exists. Add its hex ID to the map and proceed to the next record
			commenterHex[commenter.CommenterHex] = c.CommenterHexID()
			continue
		} else if err != svc.ErrNotFound {
			// Any other error than "not found"
			return 0, util.ErrorInternal
		}

		// Generate a random password string
		randomPassword, err := data.RandomHexID()
		if err != nil {
			logger.Errorf("cannot generate random password for new commenter: %v", err)
			return 0, util.ErrorInternal
		}

		// Persist a new commenter instance
		if c, err := svc.TheUserService.CreateCommenter(string(commenter.Email), commenter.Name, commenter.Link, commenter.Photo, "", string(randomPassword)); err != nil {
			return 0, err
		} else {
			// Save the new commenter's hex ID in the map
			commenterHex[commenter.CommenterHex] = c.CommenterHexID()
		}
	}

	// Create a map of (parent hex, comments)
	comments := map[models.ParentHexID][]models.Comment{}
	for _, comment := range inputData.Comments {
		comments[comment.ParentHex] = append(comments[comment.ParentHex], comment)
	}

	// Import comments, creating a map of comment hex (old hex, new hex)
	commentHex := map[models.ParentHexID]models.ParentHexID{data.RootParentHexID: data.RootParentHexID}
	numImported := 0
	keys := []models.ParentHexID{data.RootParentHexID}
	for i := 0; i < len(keys); i++ {
		for _, comment := range comments[keys[i]] {
			cHex, ok := commenterHex[comment.CommenterHex]
			if !ok {
				logger.Errorf("cannot get commenter: %v", err)
				return numImported, util.ErrorInternal
			}
			parentHex, ok := commentHex[comment.ParentHex]
			if !ok {
				logger.Errorf("cannot get parent comment: %v", err)
				return numImported, util.ErrorInternal
			}

			newComment, err := svc.TheCommentService.Create(
				cHex, domain, comment.URL, comment.Markdown, parentHex, comment.State, comment.CreationDate)
			if err != nil {
				return numImported, err
			}
			commentHex[models.ParentHexID(comment.CommentHex)] = models.ParentHexID(newComment.CommentHex)
			numImported++
			keys = append(keys, models.ParentHexID(comment.CommentHex))
		}
	}

	return numImported, nil
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

var pathMatch = regexp.MustCompile(`(https?://[^/]*)`)

func pathStrip(url string) string {
	return pathMatch.ReplaceAllString(url, "")
}

func domainImportDisqus(domain string, url string) (int, error) {
	if domain == "" || url == "" {
		return 0, util.ErrorMissingField
	}

	// TODO: make sure this is from disqus.com
	resp, err := http.Get(url)
	if err != nil {
		logger.Errorf("cannot get url: %v", err)
		return 0, util.ErrorCannotDownloadDisqus
	}

	defer resp.Body.Close()

	zr, err := gzip.NewReader(resp.Body)
	if err != nil {
		logger.Errorf("cannot create gzip reader: %v", err)
		return 0, util.ErrorInternal
	}

	contents, err := io.ReadAll(zr)
	if err != nil {
		logger.Errorf("cannot read gzip contents uncompressed: %v", err)
		return 0, util.ErrorInternal
	}

	x := disqusXML{}
	err = xml.Unmarshal(contents, &x)
	if err != nil {
		logger.Errorf("cannot unmarshal XML: %v", err)
		return 0, util.ErrorInternal
	}

	// Map Disqus thread IDs to threads.
	threads := make(map[string]disqusThread)
	for _, thread := range x.Threads {
		threads[thread.Id] = thread
	}

	// Map Disqus emails to commenterHex (if not available, create a new one with a random password that can be reset
	// later)
	commenterHex := map[strfmt.Email]models.CommenterHexID{}
	for _, post := range x.Posts {
		if post.IsDeleted || post.IsSpam {
			continue
		}

		// Skip authors whose email has already been processed
		email := strfmt.Email(post.Author.Username + "@disqus.com")
		if _, ok := commenterHex[email]; ok {
			continue
		}

		// Try to find an existing commenter with this email
		if c, err := svc.TheUserService.FindCommenterByIdPEmail("", string(email), false); err == nil {
			// Commenter already exists. Add its hex ID to the map and proceed to the next record
			commenterHex[email] = c.CommenterHexID()
			continue
		} else if err != svc.ErrNotFound {
			// Any other error than "not found"
			return 0, util.ErrorInternal
		}

		// Generate a random password string
		randomPassword, err := data.RandomHexID()
		if err != nil {
			logger.Errorf("cannot generate random password for new commenter: %v", err)
			return 0, util.ErrorInternal
		}

		if c, err := svc.TheUserService.CreateCommenter(string(email), post.Author.Name, "undefined", "undefined", "", string(randomPassword)); err != nil {
			return 0, err
		} else {
			commenterHex[email] = c.CommenterHexID()
		}
	}

	// For each Disqus post, create a Comentario comment. Attempt to convert the
	// HTML to markdown.
	numImported := 0
	disqusIdMap := map[string]models.HexID{}
	for _, post := range x.Posts {
		if post.IsDeleted || post.IsSpam {
			continue
		}

		cHex := data.AnonymousCommenterHexID
		if !post.Author.IsAnonymous {
			cHex = commenterHex[strfmt.Email(post.Author.Username+"@disqus.com")]
		}

		parentHex := data.RootParentHexID
		if val, ok := disqusIdMap[post.ParentId.Id]; ok {
			parentHex = models.ParentHexID(val)
		}

		// TODO: restrict the list of tags to just the basics: <a>, <b>, <i>, <code>
		// Especially remove <img> (convert it to <a>).
		comment, err := svc.TheCommentService.Create(
			cHex,
			domain,
			pathStrip(threads[post.ThreadId.Id].URL),
			html2md.Convert(post.Message),
			parentHex,
			models.CommentStateApproved,
			strfmt.DateTime(post.CreationDate))
		if err != nil {
			return numImported, err
		}

		disqusIdMap[post.Id] = comment.CommentHex
		numImported += 1
	}

	return numImported, nil
}
