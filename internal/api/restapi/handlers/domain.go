package handlers

import (
	"fmt"
	"github.com/go-openapi/runtime/middleware"
	"gitlab.com/comentario/comentario/internal/api/models"
	"gitlab.com/comentario/comentario/internal/api/restapi/operations"
	"gitlab.com/comentario/comentario/internal/svc"
	"gitlab.com/comentario/comentario/internal/util"
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

	// Succeeded
	return operations.NewDomainListOK().WithPayload(&operations.DomainListOKBody{
		ConfiguredOauths: &operations.DomainListOKBodyConfiguredOauths{
			Github: githubConfigured,
			Gitlab: gitlabConfigured,
			Google: googleConfigured,
		},
		Domains: domains,
		Success: true,
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
	return s.Scan(
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
		&d.CommentoProvider,
		&d.GoogleProvider,
		&d.GithubProvider,
		&d.GitlabProvider,
		&d.SsoProvider,
		&d.SsoSecret,
		&d.SsoURL,
		&d.DefaultSortPolicy,
	)
}

func domainUpdate(d *models.Domain) error {
	if d.SsoProvider && d.SsoURL == "" {
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

	_, err := svc.DB.Exec(statement,
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
		d.SsoURL,
		d.DefaultSortPolicy)
	if err != nil {
		logger.Errorf("cannot update non-moderators: %v", err)
		return util.ErrorInternal
	}
	return nil
}

func isDomainModerator(domain string, email string) (bool, error) {
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
