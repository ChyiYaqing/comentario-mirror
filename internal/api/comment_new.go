package api

import (
	"gitlab.com/commento/commento/api/internal/mail"
	"gitlab.com/commento/commento/api/internal/util"
	"net/http"
	"time"
)

// Take `creationDate` as a param because comment import (from Disqus, for
// example) will require a custom time.
func commentNew(commenterHex string, domain string, path string, parentHex string, markdown string, state string, creationDate time.Time) (string, error) {
	// path is allowed to be empty
	if commenterHex == "" || domain == "" || parentHex == "" || markdown == "" || state == "" {
		return "", util.ErrorMissingField
	}

	p, err := pageGet(domain, path)
	if err != nil {
		logger.Errorf("cannot get page attributes: %v", err)
		return "", util.ErrorInternal
	}

	if p.IsLocked {
		return "", util.ErrorThreadLocked
	}

	commentHex, err := util.RandomHex(32)
	if err != nil {
		return "", err
	}

	html := markdownToHtml(markdown)

	if err = pageNew(domain, path); err != nil {
		return "", err
	}

	statement := `
		insert into comments(commentHex, domain, path, commenterHex, parentHex, markdown, html, creationDate, state)
			values($1, $2, $3, $4, $5, $6, $7, $8, $9);
	`
	_, err = DB.Exec(statement, commentHex, domain, path, commenterHex, parentHex, markdown, html, creationDate, state)
	if err != nil {
		logger.Errorf("cannot insert comment: %v", err)
		return "", util.ErrorInternal
	}

	return commentHex, nil
}

func commentNewHandler(w http.ResponseWriter, r *http.Request) {
	type request struct {
		CommenterToken *string `json:"commenterToken"`
		Domain         *string `json:"domain"`
		Path           *string `json:"path"`
		ParentHex      *string `json:"parentHex"`
		Markdown       *string `json:"markdown"`
	}

	var x request
	if err := BodyUnmarshal(r, &x); err != nil {
		BodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	domain := domainStrip(*x.Domain)
	path := *x.Path

	d, err := domainGet(domain)
	if err != nil {
		BodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	if d.State == "frozen" {
		BodyMarshalChecked(w, response{"success": false, "message": util.ErrorDomainFrozen.Error()})
		return
	}

	if d.RequireIdentification && *x.CommenterToken == "anonymous" {
		BodyMarshalChecked(w, response{"success": false, "message": util.ErrorNotAuthorised.Error()})
		return
	}

	var commenterHex, commenterName, commenterEmail, commenterLink string
	var isModerator bool
	if *x.CommenterToken == "anonymous" {
		commenterHex, commenterName, commenterEmail, commenterLink = "anonymous", "Anonymous", "", ""
	} else {
		c, err := commenterGetByCommenterToken(*x.CommenterToken)
		if err != nil {
			BodyMarshalChecked(w, response{"success": false, "message": err.Error()})
			return
		}
		commenterHex, commenterName, commenterEmail, commenterLink = c.CommenterHex, c.Name, c.Email, c.Link
		for _, mod := range d.Moderators {
			if mod.Email == c.Email {
				isModerator = true
				break
			}
		}
	}

	var state string
	if isModerator {
		state = "approved"
	} else if d.RequireModeration {
		state = "unapproved"
	} else if commenterHex == "anonymous" && d.ModerateAllAnonymous {
		state = "unapproved"
	} else if d.AutoSpamFilter && isSpam(*x.Domain, getIP(r), getUserAgent(r), commenterName, commenterEmail, commenterLink, *x.Markdown) {
		state = "flagged"
	} else {
		state = "approved"
	}

	commentHex, err := commentNew(commenterHex, domain, path, *x.ParentHex, *x.Markdown, state, time.Now().UTC())
	if err != nil {
		BodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	// TODO: reuse html in commentNew and do only one markdown to HTML conversion?
	html := markdownToHtml(*x.Markdown)

	BodyMarshalChecked(w, response{"success": true, "commentHex": commentHex, "state": state, "html": html})
	if mail.SMTPConfigured {
		go emailNotificationNew(d, path, commenterHex, commentHex, html, *x.ParentHex, state)
	}
}
