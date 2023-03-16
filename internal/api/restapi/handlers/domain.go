package handlers

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/markbates/goth"
	"gitlab.com/comentario/comentario/internal/api/exmodels"
	"gitlab.com/comentario/comentario/internal/api/models"
	"gitlab.com/comentario/comentario/internal/api/restapi/operations"
	"gitlab.com/comentario/comentario/internal/data"
	"gitlab.com/comentario/comentario/internal/svc"
	"gitlab.com/comentario/comentario/internal/util"
	"net/url"
	"strings"
)

func DomainClear(params operations.DomainClearParams) middleware.Responder {
	user, err := svc.TheUserService.FindOwnerByToken(*params.Body.OwnerToken)
	if err != nil {
		return respServiceError(err)
	}

	isOwner, err := domainOwnershipVerify(user.HexID, *params.Body.Domain)
	if err != nil {
		return operations.NewDomainClearOK().WithPayload(&models.APIResponseBase{Message: err.Error()})
	}
	if !isOwner {
		return operations.NewDomainClearOK().WithPayload(&models.APIResponseBase{Message: util.ErrorNotAuthorised.Error()})
	}

	// Clear all domain's pages/comments/votes
	if err = svc.TheDomainService.Clear(*params.Body.Domain); err != nil {
		return respServiceError(err)
	}

	// Succeeded
	return operations.NewDomainClearOK().WithPayload(&models.APIResponseBase{Success: true})
}

func DomainDelete(params operations.DomainDeleteParams) middleware.Responder {
	user, err := svc.TheUserService.FindOwnerByToken(*params.Body.OwnerToken)
	if err != nil {
		return respServiceError(err)
	}

	// Verify domain ownership
	isOwner, err := domainOwnershipVerify(user.HexID, *params.Body.Domain)
	if err != nil {
		return operations.NewDomainDeleteOK().WithPayload(&models.APIResponseBase{Message: err.Error()})
	}
	if !isOwner {
		return operations.NewDomainDeleteOK().WithPayload(&models.APIResponseBase{Message: util.ErrorNotAuthorised.Error()})
	}

	// Delete the domain
	if err = svc.TheDomainService.Delete(*params.Body.Domain); err != nil {
		return respServiceError(err)
	}

	// Succeeded
	return operations.NewDomainDeleteOK().WithPayload(&models.APIResponseBase{Success: true})
}

func DomainList(_ operations.DomainListParams, user *data.User) middleware.Responder {
	// Fetch domains by the owner
	domains, err := svc.TheDomainService.ListByOwner(user.HexID)
	if err != nil {
		return respServiceError(err)
	}

	// Prepare an IdentityProviderMap
	idps := exmodels.IdentityProviderMap{}
	for idp, gothIdP := range util.FederatedIdProviders {
		idps[idp] = goth.GetProviders()[gothIdP] != nil
	}

	// Succeeded
	return operations.NewDomainListOK().WithPayload(&operations.DomainListOKBody{
		ConfiguredOauths: idps,
		Domains:          domains,
		Success:          true,
	})
}

func DomainModeratorDelete(params operations.DomainModeratorDeleteParams) middleware.Responder {
	user, err := svc.TheUserService.FindOwnerByToken(*params.Body.OwnerToken)
	if err != nil {
		return respServiceError(err)
	}

	domainName := data.TrimmedString(params.Body.Domain)
	authorised, err := domainOwnershipVerify(user.HexID, domainName)
	if err != nil {
		return operations.NewDomainModeratorDeleteOK().WithPayload(&models.APIResponseBase{Message: err.Error()})
	}
	if !authorised {
		return operations.NewDomainModeratorDeleteOK().WithPayload(&models.APIResponseBase{Message: util.ErrorNotAuthorised.Error()})
	}

	// Delete the moderator from the database
	if err = svc.TheDomainService.DeleteModerator(domainName, data.EmailToString(params.Body.Email)); err != nil {
		return respServiceError(err)
	}

	// Succeeded
	return operations.NewDomainModeratorDeleteOK().WithPayload(&models.APIResponseBase{Success: true})
}

func DomainModeratorNew(params operations.DomainModeratorNewParams) middleware.Responder {
	user, err := svc.TheUserService.FindOwnerByToken(*params.Body.OwnerToken)
	if err != nil {
		return respServiceError(err)
	}

	domainName := *params.Body.Domain
	isOwner, err := domainOwnershipVerify(user.HexID, domainName)
	if err != nil {
		return operations.NewDomainModeratorNewOK().WithPayload(&models.APIResponseBase{Message: err.Error()})
	}
	if !isOwner {
		return operations.NewDomainModeratorNewOK().WithPayload(&models.APIResponseBase{Message: util.ErrorNotAuthorised.Error()})
	}

	// Register a new domain moderator
	if _, err := svc.TheDomainService.CreateModerator(domainName, data.EmailToString(params.Body.Email)); err != nil {
		return respServiceError(err)
	}

	// Succeeded
	return operations.NewDomainModeratorNewOK().WithPayload(&models.APIResponseBase{Success: true})
}

func DomainNew(params operations.DomainNewParams) middleware.Responder {
	user, err := svc.TheUserService.FindOwnerByToken(*params.Body.OwnerToken)
	if err != nil {
		return respServiceError(err)
	}

	// If the domain name contains a non-hostname char, parse the passed domain as a URL to only keep the host part
	domainName := data.TrimmedString(params.Body.Domain)
	if strings.ContainsAny(domainName, "/:?&") {
		if u, err := url.Parse(domainName); err != nil {
			logger.Warningf("DomainNew(): url.Parse() failed for '%s': %v", domainName, err)
			return respBadRequest(util.ErrorInvalidDomainURL)
		} else if u.Host == "" {
			logger.Warningf("DomainNew(): '%s' parses into an empty host", domainName)
			return respBadRequest(util.ErrorInvalidDomainURL)
		} else {
			// Domain can be 'host' or 'host:port'
			domainName = u.Host
		}
	}

	// Validate what's left
	if ok, _, _ := util.IsValidHostPort(domainName); !ok {
		logger.Warningf("DomainNew(): '%s' is not a valid host[:port]", domainName)
		return respBadRequest(util.ErrorInvalidDomainHost)
	}

	// Persist a new domain record in the database
	domain, err := svc.TheDomainService.Create(user.HexID, data.TrimmedString(params.Body.Name), domainName)
	if err != nil {
		return respServiceError(err)
	}

	// Register the current owner as a domain moderator
	if _, err := svc.TheDomainService.CreateModerator(domain.Domain, user.Email); err != nil {
		return respServiceError(err)
	}

	// Succeeded
	return operations.NewDomainNewOK().WithPayload(&operations.DomainNewOKBody{
		Domain:  domain.Domain,
		Success: true,
	})
}

func DomainSsoSecretNew(params operations.DomainSsoSecretNewParams) middleware.Responder {
	user, err := svc.TheUserService.FindOwnerByToken(*params.Body.OwnerToken)
	if err != nil {
		return respServiceError(err)
	}

	domainName := *params.Body.Domain
	isOwner, err := domainOwnershipVerify(user.HexID, domainName)
	if err != nil {
		return operations.NewDomainSsoSecretNewOK().WithPayload(&operations.DomainSsoSecretNewOKBody{Message: err.Error()})
	}
	if !isOwner {
		return operations.NewDomainSsoSecretNewOK().WithPayload(&operations.DomainSsoSecretNewOKBody{Message: util.ErrorNotAuthorised.Error()})
	}

	ssoSecret, err := domainSsoSecretNew(domainName)
	if err != nil {
		return operations.NewDomainSsoSecretNewOK().WithPayload(&operations.DomainSsoSecretNewOKBody{Message: err.Error()})
	}

	// Succeeded
	return operations.NewDomainSsoSecretNewOK().WithPayload(&operations.DomainSsoSecretNewOKBody{
		SsoSecret: ssoSecret,
		Success:   true,
	})
}

func DomainStatistics(params operations.DomainStatisticsParams) middleware.Responder {
	user, err := svc.TheUserService.FindOwnerByToken(*params.Body.OwnerToken)
	if err != nil {
		return respServiceError(err)
	}

	domainName := *params.Body.Domain
	isOwner, err := domainOwnershipVerify(user.HexID, domainName)
	if err != nil {
		return operations.NewDomainStatisticsOK().WithPayload(&operations.DomainStatisticsOKBody{Message: err.Error()})
	}
	if !isOwner {
		return operations.NewDomainStatisticsOK().WithPayload(&operations.DomainStatisticsOKBody{Message: util.ErrorNotAuthorised.Error()})
	}

	viewsLast30Days, err := domainStatistics(domainName)
	if err != nil {
		return operations.NewDomainStatisticsOK().WithPayload(&operations.DomainStatisticsOKBody{Message: err.Error()})
	}

	commentsLast30Days, err := commentStatistics(domainName)
	if err != nil {
		return operations.NewDomainStatisticsOK().WithPayload(&operations.DomainStatisticsOKBody{Message: err.Error()})
	}

	// Succeeded
	return operations.NewDomainStatisticsOK().WithPayload(&operations.DomainStatisticsOKBody{
		CommentsLast30Days: commentsLast30Days,
		Success:            true,
		ViewsLast30Days:    viewsLast30Days,
	})
}

func DomainUpdate(params operations.DomainUpdateParams) middleware.Responder {
	user, err := svc.TheUserService.FindOwnerByToken(*params.Body.OwnerToken)
	if err != nil {
		return respServiceError(err)
	}

	isOwner, err := domainOwnershipVerify(user.HexID, params.Body.Domain.Domain)
	if err != nil {
		return operations.NewDomainUpdateOK().WithPayload(&models.APIResponseBase{Message: err.Error()})
	}
	if !isOwner {
		return operations.NewDomainUpdateOK().WithPayload(&models.APIResponseBase{Message: util.ErrorNotAuthorised.Error()})
	}

	if err = domainUpdate(params.Body.Domain); err != nil {
		return operations.NewDomainUpdateOK().WithPayload(&models.APIResponseBase{Message: err.Error()})
	}

	// Succeeded
	return operations.NewDomainUpdateOK().WithPayload(&models.APIResponseBase{Success: true})
}

func commentStatistics(domain string) ([]int64, error) {
	statement := `
		select COUNT(comments.creationDate)
		from (
			select to_char(date_trunc('day', (current_date - offs)), 'YYYY-MM-DD') as date
			from generate_series(0, 30, 1) as offs
		) gen 
		    left outer join comments
			on 
				gen.date = to_char(date_trunc('day', comments.creationDate), 'YYYY-MM-DD') and
				comments.domain=$1
		group by gen.date
		order by gen.date;
	`
	rows, err := svc.DB.Query(statement, domain)
	if err != nil {
		logger.Errorf("cannot get daily views: %v", err)
		return nil, util.ErrorInternal
	}

	defer rows.Close()

	var last30Days []int64
	for rows.Next() {
		var count int64
		if err = rows.Scan(&count); err != nil {
			logger.Errorf("cannot get daily comments for the last month: %v", err)
			return nil, util.ErrorInternal
		}
		last30Days = append(last30Days, count)
	}

	return last30Days, nil
}

func domainOwnershipVerify(ownerHex models.HexID, domain string) (bool, error) {
	if ownerHex == "" || domain == "" {
		return false, util.ErrorMissingField
	}

	row := svc.DB.QueryRow("select exists(select 1 from domains where ownerHex=$1 and domain=$2);", ownerHex, domain)
	var exists bool
	if err := row.Scan(&exists); err != nil {
		logger.Errorf("cannot query if domain owner: %v", err)
		return false, util.ErrorInternal
	}

	return exists, nil
}

func domainSsoSecretNew(domain string) (models.HexID, error) {
	if domain == "" {
		return "", util.ErrorMissingField
	}

	ssoSecret, err := data.RandomHexID()
	if err != nil {
		logger.Errorf("error generating SSO secret hex: %v", err)
		return "", util.ErrorInternal
	}

	if err = svc.DB.Exec("update domains set ssoSecret = $2 where domain = $1;", domain, ssoSecret); err != nil {
		logger.Errorf("cannot update ssoSecret: %v", err)
		return "", util.ErrorInternal
	}

	return ssoSecret, nil
}

func domainStatistics(domain string) ([]int64, error) {
	statement := `
		select COUNT(views.viewDate)
		from (
			select to_char(date_trunc('day', (current_date - offs)), 'YYYY-MM-DD') as date
			from generate_series(0, 30, 1) as offs
		) gen left outer join views
		on gen.date = to_char(date_trunc('day', views.viewDate), 'YYYY-MM-DD') and
		   views.domain=$1
		group by gen.date
		order by gen.date;
	`
	rows, err := svc.DB.Query(statement, domain)
	if err != nil {
		logger.Errorf("cannot get daily views: %v", err)
		return nil, util.ErrorInternal
	}

	defer rows.Close()

	var last30Days []int64
	for rows.Next() {
		var count int64
		if err = rows.Scan(&count); err != nil {
			logger.Errorf("cannot get daily views for the last month: %v", err)
			return nil, util.ErrorInternal
		}
		last30Days = append(last30Days, count)
	}

	return last30Days, nil
}

func domainUpdate(d *models.Domain) error {
	if d.Idps["sso"] && d.SsoURL == "" {
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
			twitterProvider=$13,
			ssoProvider=$14,
			ssoUrl=$15,
			defaultSortPolicy=$16
		where domain=$1;
	`

	err := svc.DB.Exec(statement,
		d.Domain,
		d.Name,
		d.State,
		d.AutoSpamFilter,
		d.RequireModeration,
		d.RequireIdentification,
		d.ModerateAllAnonymous,
		d.EmailNotificationPolicy,
		d.Idps["commento"],
		d.Idps["google"],
		d.Idps["github"],
		d.Idps["gitlab"],
		d.Idps["twitter"],
		d.Idps["sso"],
		d.SsoURL,
		d.DefaultSortPolicy)
	if err != nil {
		logger.Errorf("cannot update non-moderators: %v", err)
		return util.ErrorInternal
	}
	return nil
}
