package api

import (
	"gitlab.com/commento/commento/api/internal/util"
	"net/http"
)

func domainUpdate(d domain) error {
	if d.SsoProvider && d.SsoUrl == "" {
		return util.ErrorMissingField
	}

	statement := `
		update domains
		set
			name=$2,
			state=$3,
			autoSpamFilter=$4,
			requireModeration=$5,
			requireIdentification=$6,
			moderateAllAnonymous=$7,
			emailNotificationPolicy=$8,
			commentoProvider=$9,
			googleProvider=$10,
			githubProvider=$11,
			gitlabProvider=$12,
			ssoProvider=$13,
			ssoUrl=$14,
			defaultSortPolicy=$15
		where domain=$1;
	`

	_, err := DB.Exec(statement,
		d.Domain,
		d.Name,
		d.State,
		d.AutoSpamFilter,
		d.RequireModeration,
		d.RequireIdentification,
		d.ModerateAllAnonymous,
		d.EmailNotificationPolicy,
		d.CommentoProvider,
		d.GoogleProvider,
		d.GithubProvider,
		d.GitlabProvider,
		d.SsoProvider,
		d.SsoUrl,
		d.DefaultSortPolicy)
	if err != nil {
		logger.Errorf("cannot update non-moderators: %v", err)
		return util.ErrorInternal
	}

	return nil
}

func domainUpdateHandler(w http.ResponseWriter, r *http.Request) {
	type request struct {
		OwnerToken *string `json:"ownerToken"`
		D          *domain `json:"domain"`
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

	domain := domainStrip((*x.D).Domain)
	isOwner, err := domainOwnershipVerify(o.OwnerHex, domain)
	if err != nil {
		BodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	if !isOwner {
		BodyMarshalChecked(w, response{"success": false, "message": util.ErrorNotAuthorised.Error()})
		return
	}

	if err = domainUpdate(*x.D); err != nil {
		BodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	BodyMarshalChecked(w, response{"success": true})
}
