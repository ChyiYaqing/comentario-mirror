package api

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"gitlab.com/commento/commento/api/internal/util"
	"io"
	"net/http"
)

type commentoExportV1 struct {
	Version    int         `json:"version"`
	Comments   []comment   `json:"comments"`
	Commenters []commenter `json:"commenters"`
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
	commenterHex := map[string]string{"anonymous": "anonymous"}
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
	comments := make(map[string][]comment)
	for _, comment := range data.Comments {
		parentHex := comment.ParentHex
		comments[parentHex] = append(comments[parentHex], comment)
	}

	// Import comments, creating a map of comment hex (old hex, new hex)
	commentHex := map[string]string{"root": "root"}
	numImported := 0
	keys := []string{"root"}
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

			hex, err := commentNew(
				cHex,
				domain,
				comment.Path,
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

func domainImportCommentoHandler(w http.ResponseWriter, r *http.Request) {
	type request struct {
		OwnerToken *string `json:"ownerToken"`
		Domain     *string `json:"domain"`
		URL        *string `json:"url"`
	}

	var x request
	if err := BodyUnmarshal(r, &x); err != nil {
		BodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	o, err := ownerGetByOwnerToken(*x.OwnerToken)
	if err != nil {
		BodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	domain := domainStrip(*x.Domain)
	isOwner, err := domainOwnershipVerify(o.OwnerHex, domain)
	if err != nil {
		BodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	if !isOwner {
		BodyMarshalChecked(w, response{"success": false, "message": util.ErrorNotAuthorised.Error()})
		return
	}

	numImported, err := domainImportCommento(domain, *x.URL)
	if err != nil {
		BodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	BodyMarshalChecked(w, response{"success": true, "numImported": numImported})
}
