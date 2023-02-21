package handlers

import (
	"fmt"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/markbates/goth"
	"gitlab.com/comentario/comentario/internal/api/exmodels"
	"gitlab.com/comentario/comentario/internal/api/models"
	"gitlab.com/comentario/comentario/internal/api/restapi/operations"
	"gitlab.com/comentario/comentario/internal/svc"
	"gitlab.com/comentario/comentario/internal/util"
	"strings"
	"time"
)

var domainsRowColumns = `
	domains.domain,
	domains.ownerHex,
	domains.name,
	domains.creationDate,
	domains.state,
	domains.importedComments,
	domains.autoSpamFilter,
	domains.requireModeration,
	domains.requireIdentification,
	domains.moderateAllAnonymous,
	domains.emailNotificationPolicy,
	domains.commentoProvider,
	domains.googleProvider,
	domains.githubProvider,
	domains.gitlabProvider,
	domains.twitterProvider,
	domains.ssoProvider,
	domains.ssoSecret,
	domains.ssoUrl,
	domains.defaultSortPolicy
`

func DomainClear(params operations.DomainClearParams) middleware.Responder {
	owner, err := ownerGetByOwnerToken(*params.Body.OwnerToken)
	if err != nil {
		return operations.NewDomainClearOK().WithPayload(&models.APIResponseBase{Message: err.Error()})
	}

	isOwner, err := domainOwnershipVerify(owner.OwnerHex, *params.Body.Domain)
	if err != nil {
		return operations.NewDomainClearOK().WithPayload(&models.APIResponseBase{Message: err.Error()})
	}
	if !isOwner {
		return operations.NewDomainClearOK().WithPayload(&models.APIResponseBase{Message: util.ErrorNotAuthorised.Error()})
	}

	if err = domainClear(*params.Body.Domain); err != nil {
		return operations.NewDomainClearOK().WithPayload(&models.APIResponseBase{Message: err.Error()})
	}

	// Succeeded
	return operations.NewDomainClearOK().WithPayload(&models.APIResponseBase{Success: true})
}

func DomainDelete(params operations.DomainDeleteParams) middleware.Responder {
	owner, err := ownerGetByOwnerToken(*params.Body.OwnerToken)
	if err != nil {
		return operations.NewDomainDeleteOK().WithPayload(&models.APIResponseBase{Message: err.Error()})
	}

	isOwner, err := domainOwnershipVerify(owner.OwnerHex, *params.Body.Domain)
	if err != nil {
		return operations.NewDomainDeleteOK().WithPayload(&models.APIResponseBase{Message: err.Error()})
	}
	if !isOwner {
		return operations.NewDomainDeleteOK().WithPayload(&models.APIResponseBase{Message: util.ErrorNotAuthorised.Error()})
	}

	if err = domainDelete(*params.Body.Domain); err != nil {
		return operations.NewDomainDeleteOK().WithPayload(&models.APIResponseBase{Message: err.Error()})
	}

	// Succeeded
	return operations.NewDomainDeleteOK().WithPayload(&models.APIResponseBase{Success: true})
}

func DomainList(params operations.DomainListParams) middleware.Responder {
	owner, err := ownerGetByOwnerToken(*params.Body.OwnerToken)
	if err != nil {
		return operations.NewDomainListOK().WithPayload(&operations.DomainListOKBody{Message: err.Error()})
	}

	domains, err := domainList(owner.OwnerHex)
	if err != nil {
		return operations.NewDomainListOK().WithPayload(&operations.DomainListOKBody{Message: err.Error()})
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
	owner, err := ownerGetByOwnerToken(*params.Body.OwnerToken)
	if err != nil {
		return operations.NewDomainModeratorDeleteOK().WithPayload(&models.APIResponseBase{Message: err.Error()})
	}

	domainName := *params.Body.Domain
	authorised, err := domainOwnershipVerify(owner.OwnerHex, domainName)
	if err != nil {
		return operations.NewDomainModeratorDeleteOK().WithPayload(&models.APIResponseBase{Message: err.Error()})
	}
	if !authorised {
		return operations.NewDomainModeratorDeleteOK().WithPayload(&models.APIResponseBase{Message: util.ErrorNotAuthorised.Error()})
	}

	if err = domainModeratorDelete(domainName, *params.Body.Email); err != nil {
		return operations.NewDomainModeratorDeleteOK().WithPayload(&models.APIResponseBase{Message: err.Error()})
	}

	// Succeeded
	return operations.NewDomainModeratorDeleteOK().WithPayload(&models.APIResponseBase{Success: true})
}

func DomainModeratorNew(params operations.DomainModeratorNewParams) middleware.Responder {
	owner, err := ownerGetByOwnerToken(*params.Body.OwnerToken)
	if err != nil {
		return operations.NewDomainModeratorNewOK().WithPayload(&models.APIResponseBase{Message: err.Error()})
	}

	domainName := *params.Body.Domain
	isOwner, err := domainOwnershipVerify(owner.OwnerHex, domainName)
	if err != nil {
		return operations.NewDomainModeratorNewOK().WithPayload(&models.APIResponseBase{Message: err.Error()})
	}
	if !isOwner {
		return operations.NewDomainModeratorNewOK().WithPayload(&models.APIResponseBase{Message: util.ErrorNotAuthorised.Error()})
	}

	if err = domainModeratorNew(domainName, *params.Body.Email); err != nil {
		return operations.NewDomainModeratorNewOK().WithPayload(&models.APIResponseBase{Message: err.Error()})
	}

	// Succeeded
	return operations.NewDomainModeratorNewOK().WithPayload(&models.APIResponseBase{Success: true})
}

func DomainNew(params operations.DomainNewParams) middleware.Responder {
	owner, err := ownerGetByOwnerToken(*params.Body.OwnerToken)
	if err != nil {
		return operations.NewDomainNewOK().WithPayload(&operations.DomainNewOKBody{Message: err.Error()})
	}

	domainName := *params.Body.Domain
	if err = domainNew(owner.OwnerHex, *params.Body.Name, domainName); err != nil {
		return operations.NewDomainNewOK().WithPayload(&operations.DomainNewOKBody{Message: err.Error()})
	}

	if err = domainModeratorNew(domainName, owner.Email); err != nil {
		return operations.NewDomainNewOK().WithPayload(&operations.DomainNewOKBody{Message: err.Error()})
	}

	// Succeeded
	return operations.NewDomainNewOK().WithPayload(&operations.DomainNewOKBody{
		Domain:  domainName,
		Success: true,
	})
}

func DomainSsoSecretNew(params operations.DomainSsoSecretNewParams) middleware.Responder {
	owner, err := ownerGetByOwnerToken(*params.Body.OwnerToken)
	if err != nil {
		return operations.NewDomainSsoSecretNewOK().WithPayload(&operations.DomainSsoSecretNewOKBody{Message: err.Error()})
	}

	domainName := *params.Body.Domain
	isOwner, err := domainOwnershipVerify(owner.OwnerHex, domainName)
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
	owner, err := ownerGetByOwnerToken(*params.Body.OwnerToken)
	if err != nil {
		return operations.NewDomainStatisticsOK().WithPayload(&operations.DomainStatisticsOKBody{Message: err.Error()})
	}

	domainName := *params.Body.Domain
	isOwner, err := domainOwnershipVerify(owner.OwnerHex, domainName)
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
	owner, err := ownerGetByOwnerToken(*params.Body.OwnerToken)
	if err != nil {
		return operations.NewDomainUpdateOK().WithPayload(&models.APIResponseBase{Message: err.Error()})
	}

	isOwner, err := domainOwnershipVerify(owner.OwnerHex, params.Body.Domain.Domain)
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

func domainClear(domain string) error {
	if domain == "" {
		return util.ErrorMissingField
	}

	_, err := svc.DB.Exec("delete from votes using comments where comments.commentHex = votes.commentHex and comments.domain = $1;", domain)
	if err != nil {
		logger.Errorf("cannot delete votes: %v", err)
		return util.ErrorInternal
	}

	_, err = svc.DB.Exec("delete from comments where comments.domain = $1;", domain)
	if err != nil {
		logger.Errorf("domainClear(): DB.Exec for comments failed for domain %s: %v", domain, err)
		return util.ErrorInternal
	}

	_, err = svc.DB.Exec("delete from pages where pages.domain = $1;", domain)
	if err != nil {
		logger.Errorf("domainClear(): DB.Exec for pages failed for domain %s: %v", domain, err)
		return util.ErrorInternal
	}

	return nil
}

func domainDelete(domain string) error {
	if domain == "" {
		return util.ErrorMissingField
	}

	_, err := svc.DB.Exec("delete from domains where domain = $1;", domain)
	if err != nil {
		return util.ErrorNoSuchDomain
	}

	_, err = svc.DB.Exec("delete from views where views.domain = $1;", domain)
	if err != nil {
		logger.Errorf("cannot delete domain from views: %v", err)
		return util.ErrorInternal
	}

	_, err = svc.DB.Exec("delete from moderators where moderators.domain = $1;", domain)
	if err != nil {
		logger.Errorf("cannot delete domain from moderators: %v", err)
		return util.ErrorInternal
	}

	_, err = svc.DB.Exec("delete from ssotokens where ssotokens.domain = $1;", domain)
	if err != nil {
		logger.Errorf("cannot delete domain from ssotokens: %v", err)
		return util.ErrorInternal
	}

	// comments, votes, and pages are handled by domainClear
	if err = domainClear(domain); err != nil {
		logger.Errorf("cannot clear domain: %v", err)
		return util.ErrorInternal
	}

	return nil
}

func domainGet(dmn string) (*models.Domain, error) {
	if dmn == "" {
		return nil, util.ErrorMissingField
	}

	row := svc.DB.QueryRow(fmt.Sprintf("select %s from domains where domain = $1;", domainsRowColumns), dmn)
	var err error
	var d models.Domain
	if err = domainsRowScan(row, &d); err != nil {
		return nil, util.ErrorNoSuchDomain
	}

	d.Moderators, err = domainModeratorList(d.Domain)
	if err != nil {
		return nil, err
	}

	return &d, nil
}

func domainList(ownerHex models.HexID) ([]*models.Domain, error) {
	if ownerHex == "" {
		return nil, util.ErrorMissingField
	}

	rows, err := svc.DB.Query(
		fmt.Sprintf("select %s from domains where ownerHex=$1;", domainsRowColumns),
		ownerHex)
	if err != nil {
		logger.Errorf("cannot query domains: %v", err)
		return nil, util.ErrorInternal
	}
	defer rows.Close()

	var domains []*models.Domain
	for rows.Next() {
		var d models.Domain
		if err = domainsRowScan(rows, &d); err != nil {
			logger.Errorf("cannot Scan domain: %v", err)
			return nil, util.ErrorInternal
		}
		if d.Moderators, err = domainModeratorList(d.Domain); err != nil {
			return nil, err
		}
		domains = append(domains, &d)
	}
	return domains, rows.Err()
}

func domainModeratorDelete(domain string, email strfmt.Email) error {
	if domain == "" || email == "" {
		return util.ErrorMissingConfig
	}

	_, err := svc.DB.Exec("delete from moderators where domain=$1 and email=$2;", domain, email)
	if err != nil {
		logger.Errorf("cannot delete moderator: %v", err)
		return util.ErrorInternal
	}

	return nil
}

func domainModeratorList(domain string) ([]*models.DomainModerator, error) {
	statement := `
		select email, addDate
		from moderators
		where domain=$1;
	`
	rows, err := svc.DB.Query(statement, domain)
	if err != nil {
		logger.Errorf("cannot get moderators: %v", err)
		return nil, util.ErrorInternal
	}
	defer rows.Close()

	var moderators []*models.DomainModerator
	for rows.Next() {
		m := models.DomainModerator{}
		if err = rows.Scan(&m.Email, &m.AddDate); err != nil {
			logger.Errorf("cannot Scan moderator: %v", err)
			return nil, util.ErrorInternal
		}
		moderators = append(moderators, &m)
	}
	return moderators, nil
}

func domainModeratorNew(domain string, email strfmt.Email) error {
	if domain == "" || email == "" {
		return util.ErrorMissingField
	}

	if err := EmailNew(email); err != nil {
		logger.Errorf("cannot create email when creating moderator: %v", err)
		return util.ErrorInternal
	}

	statement := `insert into moderators(domain, email, addDate) values($1, $2, $3);`
	_, err := svc.DB.Exec(statement, domain, email, time.Now().UTC())
	if err != nil {
		logger.Errorf("cannot insert new moderator: %v", err)
		return util.ErrorInternal
	}

	return nil
}

func domainNew(ownerHex models.HexID, name string, domain string) error {
	if ownerHex == "" || name == "" || domain == "" {
		return util.ErrorMissingField
	}

	if strings.Contains(domain, "/") {
		return util.ErrorInvalidDomain
	}

	_, err := svc.DB.Exec(
		"insert into domains(ownerHex, name, domain, creationDate) values($1, $2, $3, $4);",
		ownerHex,
		name,
		domain,
		time.Now().UTC())
	if err != nil {
		// TODO: Make sure this is really the error.
		return util.ErrorDomainAlreadyExists
	}
	return nil
}

func domainOwnershipVerify(ownerHex models.HexID, domain string) (bool, error) {
	if ownerHex == "" || domain == "" {
		return false, util.ErrorMissingField
	}

	statement := `select EXISTS (select 1 from domains where ownerHex=$1 and domain=$2);`
	row := svc.DB.QueryRow(statement, ownerHex, domain)
	var exists bool
	if err := row.Scan(&exists); err != nil {
		logger.Errorf("cannot query if domain owner: %v", err)
		return false, util.ErrorInternal
	}

	return exists, nil
}

func domainsRowScan(s util.Scanner, d *models.Domain) error {
	var commento, google, github, gitlab, twitter, sso bool
	err := s.Scan(
		&d.Domain,
		&d.OwnerHex,
		&d.Name,
		&d.CreationDate,
		&d.State,
		&d.ImportedComments,
		&d.AutoSpamFilter,
		&d.RequireModeration,
		&d.RequireIdentification,
		&d.ModerateAllAnonymous,
		&d.EmailNotificationPolicy,
		&commento,
		&google,
		&github,
		&gitlab,
		&twitter,
		&sso,
		&d.SsoSecret,
		&d.SsoURL,
		&d.DefaultSortPolicy,
	)
	if err != nil {
		return err
	}

	// Compile a map of identity providers
	d.Idps = exmodels.IdentityProviderMap{
		"commento": commento,
		"google":   google,
		"github":   github,
		"gitlab":   gitlab,
		"twitter":  twitter,
		"sso":      sso,
	}
	return nil
}

func domainSsoSecretNew(domain string) (models.HexID, error) {
	if domain == "" {
		return "", util.ErrorMissingField
	}

	ssoSecret, err := util.RandomHex(32)
	if err != nil {
		logger.Errorf("error generating SSO secret hex: %v", err)
		return "", util.ErrorInternal
	}

	if _, err = svc.DB.Exec("update domains set ssoSecret = $2 where domain = $1;", domain, ssoSecret); err != nil {
		logger.Errorf("cannot update ssoSecret: %v", err)
		return "", util.ErrorInternal
	}

	return models.HexID(ssoSecret), nil
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

	_, err := svc.DB.Exec(statement,
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

func domainViewRecord(domain string, commenter *models.Commenter) {
	ch := AnonymousCommenterHexID
	if commenter != nil {
		ch = commenter.CommenterHex
	}
	_, err := svc.DB.Exec("insert into views(domain, commenterHex, viewDate) values ($1, $2, $3);", domain, ch, time.Now().UTC())
	if err != nil {
		logger.Warningf("cannot insert views: %v", err)
	}
}

func isDomainModerator(domain string, email strfmt.Email) (bool, error) {
	row := svc.DB.QueryRow(
		"select EXISTS(select 1 from moderators where domain=$1 and email=$2);",
		domain,
		email)
	var exists bool
	if err := row.Scan(&exists); err != nil {
		logger.Errorf("cannot query if moderator: %v", err)
		return false, util.ErrorInternal
	}

	return exists, nil
}
