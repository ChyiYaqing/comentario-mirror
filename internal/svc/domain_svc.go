package svc

import (
	"database/sql"
	"gitlab.com/comentario/comentario/internal/api/exmodels"
	"gitlab.com/comentario/comentario/internal/api/models"
	"time"
)

// TheDomainService is a global DomainService implementation
var TheDomainService DomainService = &domainService{}

// DomainService is a service interface for dealing with domains
type DomainService interface {
	// Clear removes all pages, comments, and comment votes for the specified domain
	Clear(domain string) error
	// Delete deletes the specified domain
	Delete(domain string) error
	// DeleteModerator deletes the specified domain moderator
	DeleteModerator(domain, email string) error
	// FindByName fetches and returns a domain with the specified name
	FindByName(domainName string) (*models.Domain, error)
	// IsDomainModerator returns whther the given email is a moderator in the given domain
	IsDomainModerator(domain, email string) (bool, error)
	// ListByOwner fetches and returns a list of domains for the specified owner
	ListByOwner(ownerHex models.HexID) ([]*models.Domain, error)
	// RegisterView records a domain view in the database. commenterHex should be "anonymous" for an unauthenticated
	// viewer
	RegisterView(domain string, commenterHex models.CommenterHexID) error
}

//----------------------------------------------------------------------------------------------------------------------

// domainService is a blueprint DomainService implementation
type domainService struct{}

func (svc *domainService) Clear(domain string) error {
	logger.Debugf("domainService.Clear(%s)", domain)

	// Remove all votes on domain's comments
	if err := TheVoteService.DeleteByDomain(domain); err != nil {
		return err
	}

	// Remove all domain's comments
	if err := TheCommentService.DeleteByDomain(domain); err != nil {
		return err
	}

	// Remove all domain's pages
	if err := ThePageService.DeleteByDomain(domain); err != nil {
		return err
	}

	// Succeeded
	return nil
}

func (svc *domainService) Delete(domain string) error {
	logger.Debugf("domainService.Delete(%s)", domain)

	// Remove the domain's view stats, moderators, ssotokens
	err := checkErrors(
		db.Exec("delete from views where domain=$1;", domain),
		db.Exec("delete from moderators where domain=$1;", domain),
		db.Exec("delete from ssotokens where domain=$1;", domain))
	if err != nil {
		logger.Errorf("domainService.Delete: Exec() failed for dependent object: %v", err)
		return translateErrors(err)
	}

	// Remove the domain itself
	if err := db.Exec("delete from domains where domain=$1;", domain); err != nil {
		logger.Errorf("domainService.Delete: Exec() failed for domain: %v", err)
		return translateErrors(err)
	}

	// Succeeded
	return nil
}

func (svc *domainService) DeleteModerator(domain, email string) error {
	logger.Debugf("domainService.DeleteModerator(%s, %s)", domain, email)

	// Remove the row from the database
	if err := db.Exec("delete from moderators where domain=$1 and email=$2;", domain, email); err != nil {
		logger.Errorf("domainService.DeleteModerator: Exec() failed: %v", err)
		return translateErrors(err)
	}

	// Succeeded
	return nil
}

func (svc *domainService) FindByName(domainName string) (*models.Domain, error) {
	logger.Debugf("domainService.Find(%s)", domainName)

	// Query the row
	rows, err := db.Query(
		"select "+
			"d.domain, d.ownerhex, d.name, d.creationdate, d.state, d.importedcomments, d.autospamfilter, "+
			"d.requiremoderation, d.requireidentification, d.moderateallanonymous, d.emailnotificationpolicy, "+
			"d.commentoprovider, d.googleprovider, d.githubprovider, d.gitlabprovider, d.twitterprovider, "+
			"d.ssoprovider, d.ssosecret, d.ssourl, d.defaultsortpolicy, m.email, m.adddate "+
			"from domains d "+
			"left join moderators m on m.domain=d.domain "+
			"where d.domain=$1;",
		domainName)
	if err != nil {
		logger.Errorf("domainService.FindByName: Query() failed: %v", err)
		return nil, translateErrors(err)
	}
	defer rows.Close()

	// Fetch the domain(s)
	if domains, err := svc.fetchDomainsAndModerators(rows); err != nil {
		return nil, translateErrors(err)
	} else if len(domains) == 0 {
		return nil, ErrNotFound
	} else {
		// Grab the first one
		return domains[0], nil
	}
}

func (svc *domainService) IsDomainModerator(domain, email string) (bool, error) {
	logger.Debugf("domainService.IsDomainModerator(%s, %s)", domain, email)

	// Query the row
	row := db.QueryRow("select 1 from moderators where domain=$1 and email=$2;", domain, email)
	var b byte
	if err := row.Scan(&b); err == sql.ErrNoRows {
		// No rows means it isn't a moderator
		return false, nil

	} else if err != nil {
		// Any other database error
		logger.Errorf("domainService.IsDomainModerator: QueryRow() failed: %v", err)
		return false, translateErrors(err)
	}

	// Succeeded: the email belongs to a domain moderator
	return true, nil
}

func (svc *domainService) ListByOwner(ownerHex models.HexID) ([]*models.Domain, error) {
	logger.Debugf("domainService.ListByOwner(%s)", ownerHex)

	// Query domains and moderators
	rows, err := db.Query(
		"select "+
			"d.domain, d.ownerhex, d.name, d.creationdate, d.state, d.importedcomments, d.autospamfilter, "+
			"d.requiremoderation, d.requireidentification, d.moderateallanonymous, d.emailnotificationpolicy, "+
			"d.commentoprovider, d.googleprovider, d.githubprovider, d.gitlabprovider, d.twitterprovider, "+
			"d.ssoprovider, d.ssosecret, d.ssourl, d.defaultsortpolicy, m.email, m.adddate "+
			"from domains d "+
			"left join moderators m on m.domain=d.domain "+
			"where d.ownerhex=$1;",
		ownerHex)
	if err != nil {
		logger.Errorf("domainService.ListByOwner: Query() failed: %v", err)
		return nil, translateErrors(err)
	}
	defer rows.Close()

	// Fetch the domains
	if domains, err := svc.fetchDomainsAndModerators(rows); err != nil {
		return nil, translateErrors(err)
	} else {
		return domains, nil
	}
}

func (svc *domainService) RegisterView(domain string, commenterHex models.CommenterHexID) error {
	logger.Debugf("domainService.RegisterView(%s, %s)", domain, commenterHex)

	err := db.Exec(
		"insert into views(domain, commenterhex, viewdate) values ($1, $2, $3);",
		domain, commenterHex, time.Now().UTC())
	if err != nil {
		logger.Warningf("domainService.RegisterView: Exec() failed: %v", err)
		return translateErrors(err)
	}

	// Succeeded
	return nil
}

// fetchDomainsAndModerators returns a list of domain instances from the provided database rows
func (svc *domainService) fetchDomainsAndModerators(rs *sql.Rows) ([]*models.Domain, error) {
	// Maintain a map of domains by name
	dn := map[string]*models.Domain{}
	var res []*models.Domain

	// Iterate all rows
	for rs.Next() {
		// Fetch a domain and a moderator
		d := models.Domain{}
		m := models.DomainModerator{}
		var commento, google, github, gitlab, twitter, sso bool
		err := rs.Scan(
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
			&m.Email,
			&m.AddDate)
		if err != nil {
			logger.Warningf("domainService.fetchDomainsAndModerators: Scan() failed: %v", err)
			return nil, err
		}

		// If the domain isn't encountered yet
		var domain *models.Domain
		var exists bool
		if domain, exists = dn[d.Domain]; !exists {
			domain = &d

			// Compile a map of identity providers
			d.Idps = exmodels.IdentityProviderMap{
				"commento": commento,
				"google":   google,
				"github":   github,
				"gitlab":   gitlab,
				"twitter":  twitter,
				"sso":      sso,
			}

			// Add the domain to the result list and the name map
			res = append(res, domain)
			dn[d.Domain] = domain
		}

		// Add the current moderator, if any, to the domain moderators
		if m.Email != "" {
			m.Domain = domain.Domain
			domain.Moderators = append(domain.Moderators, &m)
		}
	}

	// Check if Next() didn't error
	if err := rs.Err(); err != nil {
		return nil, err
	}

	// Succeeded
	return res, nil
}
