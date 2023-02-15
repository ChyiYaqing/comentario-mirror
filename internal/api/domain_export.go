package api

import (
	"encoding/json"
	"github.com/go-openapi/strfmt"
	"gitlab.com/comentario/comentario/internal/api/models"
	"gitlab.com/comentario/comentario/internal/mail"
	"gitlab.com/comentario/comentario/internal/svc"
	"gitlab.com/comentario/comentario/internal/util"
	"net/http"
	"time"
)

func domainExportBeginError(email strfmt.Email, toName string, domain string, _ error) {
	// we're not using err at the moment because it's all errorInternal
	if err2 := mail.SMTPDomainExportError(string(email), toName, domain); err2 != nil {
		logger.Errorf("cannot send domain export error email for %s: %v", domain, err2)
	}
}

func domainExportBegin(email strfmt.Email, toName string, domain string) {
	e := commentoExportV1{Version: 1, Comments: []models.Comment{}, Commenters: []commenter{}}

	statement := `
		select commentHex, domain, path, commenterHex, markdown, parentHex, score, state, creationDate
		from comments
		where domain = $1;
	`
	rows1, err := svc.DB.Query(statement, domain)
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

	statement = `
		select commenters.commenterHex, commenters.email, commenters.name, commenters.link, commenters.photo, commenters.provider, commenters.joinDate
		from commenters, comments
		where comments.domain = $1 and commenters.commenterHex = comments.commenterHex;
	`
	rows2, err := svc.DB.Query(statement, domain)
	if err != nil {
		logger.Errorf("cannot select commenters while exporting %s: %v", domain, err)
		domainExportBeginError(email, toName, domain, util.ErrorInternal)
		return
	}
	defer rows2.Close()

	for rows2.Next() {
		c := commenter{}
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

	statement = `
		insert into
		exports (exportHex, binData, domain, creationDate)
		values  ($1,        $2,      $3    , $4          );
	`
	_, err = svc.DB.Exec(statement, exportHex, gje, domain, time.Now().UTC())
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

func domainExportBeginHandler(w http.ResponseWriter, r *http.Request) {
	type request struct {
		OwnerToken *string `json:"ownerToken"`
		Domain     *string `json:"domain"`
	}

	var x request
	if err := BodyUnmarshal(r, &x); err != nil {
		BodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	if !mail.SMTPConfigured {
		BodyMarshalChecked(w, response{"success": false, "message": util.ErrorSmtpNotConfigured.Error()})
		return
	}

	o, err := OwnerGetByOwnerToken(models.HexID(*x.OwnerToken))
	if err != nil {
		BodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	isOwner, err := domainOwnershipVerify(o.OwnerHex, *x.Domain)
	if err != nil {
		BodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	if !isOwner {
		BodyMarshalChecked(w, response{"success": false, "message": util.ErrorNotAuthorised.Error()})
		return
	}

	go domainExportBegin(o.Email, o.Name, *x.Domain)

	BodyMarshalChecked(w, response{"success": true})
}
