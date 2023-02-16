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
	"gitlab.com/comentario/comentario/internal/mail"
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
	if !mail.SMTPConfigured {
		return operations.NewDomainExportBeginOK().WithPayload(&models.APIResponseBase{Message: util.ErrorSmtpNotConfigured.Error()})
	}

	owner, err := ownerGetByOwnerToken(*params.Body.OwnerToken)
	if err != nil {
		return operations.NewDomainExportBeginOK().WithPayload(&models.APIResponseBase{Message: err.Error()})
	}

	isOwner, err := domainOwnershipVerify(owner.OwnerHex, *params.Body.Domain)
	if err != nil {
		return operations.NewDomainExportBeginOK().WithPayload(&models.APIResponseBase{Message: err.Error()})
	}

	if !isOwner {
		return operations.NewDomainExportBeginOK().WithPayload(&models.APIResponseBase{Message: util.ErrorNotAuthorised.Error()})
	}

	go domainExportBegin(owner.Email, owner.Name, *params.Body.Domain)

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
	o, err := ownerGetByOwnerToken(*params.Body.OwnerToken)
	if err != nil {
		return operations.NewDomainImportCommentoOK().WithPayload(&operations.DomainImportCommentoOKBody{Message: err.Error()})
	}

	domainName := *params.Body.Domain
	isOwner, err := domainOwnershipVerify(o.OwnerHex, domainName)
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
	owner, err := ownerGetByOwnerToken(*params.Body.OwnerToken)
	if err != nil {
		return operations.NewDomainImportDisqusOK().WithPayload(&operations.DomainImportDisqusOKBody{Message: err.Error()})
	}

	domainName := *params.Body.Domain
	isOwner, err := domainOwnershipVerify(owner.OwnerHex, domainName)
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

func domainExportBeginError(email strfmt.Email, toName string, domain string, _ error) {
	// we're not using err at the moment because it's all errorInternal
	if err2 := mail.SMTPDomainExportError(string(email), toName, domain); err2 != nil {
		logger.Errorf("cannot send domain export error email for %s: %v", domain, err2)
	}
}

func domainExportBegin(email strfmt.Email, toName string, domain string) {
	e := commentoExportV1{Version: 1, Comments: []models.Comment{}, Commenters: []models.Commenter{}}

	rows1, err := svc.DB.Query(
		"select commentHex, domain, path, commenterHex, markdown, parentHex, score, state, creationDate from comments where domain = $1;",
		domain)
	if err != nil {
		logger.Errorf("cannot select comments while exporting %s: %v", domain, err)
		domainExportBeginError(email, toName, domain, util.ErrorInternal)
		return
	}
	defer rows1.Close()

	for rows1.Next() {
		c := models.Comment{}
		if err = rows1.Scan(&c.CommentHex, &c.Domain, &c.URL, &c.CommenterHex, &c.Markdown, &c.ParentHex, &c.Score, &c.State, &c.CreationDate); err != nil {
			logger.Errorf("cannot scan comment while exporting %s: %v", domain, err)
			domainExportBeginError(email, toName, domain, util.ErrorInternal)
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
		domainExportBeginError(email, toName, domain, util.ErrorInternal)
		return
	}
	defer rows2.Close()

	for rows2.Next() {
		c := models.Commenter{}
		if err := rows2.Scan(&c.CommenterHex, &c.Email, &c.Name, &c.Link, &c.Photo, &c.Provider, &c.JoinDate); err != nil {
			logger.Errorf("cannot scan commenter while exporting %s: %v", domain, err)
			domainExportBeginError(email, toName, domain, util.ErrorInternal)
			return
		}

		e.Commenters = append(e.Commenters, c)
	}

	je, err := json.Marshal(e)
	if err != nil {
		logger.Errorf("cannot marshall JSON while exporting %s: %v", domain, err)
		domainExportBeginError(email, toName, domain, util.ErrorInternal)
		return
	}

	gje, err := util.GzipStatic(je)
	if err != nil {
		logger.Errorf("cannot gzip JSON while exporting %s: %v", domain, err)
		domainExportBeginError(email, toName, domain, util.ErrorInternal)
		return
	}

	exportHex, err := util.RandomHex(32)
	if err != nil {
		logger.Errorf("cannot generate exportHex while exporting %s: %v", domain, err)
		domainExportBeginError(email, toName, domain, util.ErrorInternal)
		return
	}

	_, err = svc.DB.Exec(
		"insert into exports(exportHex, binData, domain, creationDate) values($1, $2, $3, $4);",
		exportHex,
		gje,
		domain,
		time.Now().UTC())
	if err != nil {
		logger.Errorf("error inserting expiry binary data while exporting %s: %v", domain, err)
		domainExportBeginError(email, toName, domain, util.ErrorInternal)
		return
	}

	err = mail.SMTPDomainExport(string(email), toName, domain, exportHex)
	if err != nil {
		logger.Errorf("error sending data export email for %s: %v", domain, err)
		return
	}
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

	var data commentoExportV1
	if err := json.Unmarshal(contents, &data); err != nil {
		logger.Errorf("cannot unmarshal JSON at %s: %v", url, err)
		return 0, util.ErrorInternal
	}

	if data.Version != 1 {
		logger.Errorf("invalid data version (got %d, want 1): %v", data.Version, err)
		return 0, util.ErrorUnsupportedCommentoImportVersion
	}

	// Check if imported commentedHex or email exists, creating a map of
	// commenterHex (old hex, new hex)
	commenterHex := map[models.HexID]models.HexID{"anonymous": "anonymous"}
	for _, commenter := range data.Commenters {
		c, err := commenterGetByEmail("commento", commenter.Email)
		if err != nil && err != util.ErrorNoSuchCommenter {
			logger.Errorf("cannot get commenter by email: %v", err)
			return 0, util.ErrorInternal
		}

		if err == nil {
			commenterHex[commenter.CommenterHex] = c.CommenterHex
			continue
		}

		randomPassword, err := util.RandomHex(32)
		if err != nil {
			logger.Errorf("cannot generate random password for new commenter: %v", err)
			return 0, util.ErrorInternal
		}

		commenterHex[commenter.CommenterHex], err = commenterNew(commenter.Email,
			commenter.Name, commenter.Link, commenter.Photo, "commento", randomPassword)
		if err != nil {
			return 0, err
		}
	}

	// Create a map of (parent hex, comments)
	comments := make(map[models.HexID][]models.Comment)
	for _, comment := range data.Comments {
		parentHex := models.HexID(comment.ParentHex)
		comments[parentHex] = append(comments[parentHex], comment)
	}

	// Import comments, creating a map of comment hex (old hex, new hex)
	commentHex := map[models.HexID]models.HexID{"root": "root"}
	numImported := 0
	keys := []models.HexID{"root"}
	for i := 0; i < len(keys); i++ {
		for _, comment := range comments[keys[i]] {
			cHex, ok := commenterHex[comment.CommenterHex]
			if !ok {
				logger.Errorf("cannot get commenter: %v", err)
				return numImported, util.ErrorInternal
			}
			parentHex, ok := commentHex[models.HexID(comment.ParentHex)]
			if !ok {
				logger.Errorf("cannot get parent comment: %v", err)
				return numImported, util.ErrorInternal
			}

			hex, err := commentNew(
				cHex,
				domain,
				comment.URL,
				parentHex,
				comment.Markdown,
				comment.State,
				comment.CreationDate)
			if err != nil {
				return numImported, err
			}
			commentHex[comment.CommentHex] = hex
			numImported++
			keys = append(keys, comment.CommentHex)
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

	// Map Disqus emails to commenterHex (if not available, create a new one
	// with a random password that can be reset later).
	commenterHex := map[strfmt.Email]models.HexID{}
	for _, post := range x.Posts {
		if post.IsDeleted || post.IsSpam {
			continue
		}

		email := strfmt.Email(post.Author.Username + "@disqus.com")

		if _, ok := commenterHex[email]; ok {
			continue
		}

		c, err := commenterGetByEmail("commento", email)
		if err != nil && err != util.ErrorNoSuchCommenter {
			logger.Errorf("cannot get commenter by email: %v", err)
			return 0, util.ErrorInternal
		}

		if err == nil {
			commenterHex[email] = c.CommenterHex
			continue
		}

		randomPassword, err := util.RandomHex(32)
		if err != nil {
			logger.Errorf("cannot generate random password for new commenter: %v", err)
			return 0, util.ErrorInternal
		}

		commenterHex[email], err = commenterNew(email, post.Author.Name, "undefined", "undefined", "commento", randomPassword)
		if err != nil {
			return 0, err
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

		cHex := models.HexID("anonymous")
		if !post.Author.IsAnonymous {
			cHex = commenterHex[strfmt.Email(post.Author.Username+"@disqus.com")]
		}

		parentHex := models.HexID("root")
		if val, ok := disqusIdMap[post.ParentId.Id]; ok {
			parentHex = val
		}

		// TODO: restrict the list of tags to just the basics: <a>, <b>, <i>, <code>
		// Especially remove <img> (convert it to <a>).
		commentHex, err := commentNew(
			cHex,
			domain,
			pathStrip(threads[post.ThreadId.Id].URL),
			parentHex,
			html2md.Convert(post.Message),
			"approved",
			strfmt.DateTime(post.CreationDate))
		if err != nil {
			return numImported, err
		}

		disqusIdMap[post.Id] = commentHex
		numImported += 1
	}

	return numImported, nil
}
