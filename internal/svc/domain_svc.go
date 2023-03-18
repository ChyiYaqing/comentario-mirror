package svc

import (
	"database/sql"
	"github.com/go-openapi/strfmt"
	"gitlab.com/comentario/comentario/internal/api/exmodels"
	"gitlab.com/comentario/comentario/internal/api/models"
	"gitlab.com/comentario/comentario/internal/data"
	"time"
)

// TheDomainService is a global DomainService implementation
var TheDomainService DomainService = &domainService{}

// DomainService is a service interface for dealing with domains
type DomainService interface {
	// Clear removes all pages, comments, and comment votes for the specified domain
	Clear(domain string) error
	// Create creates and persists a new domain record
	Create(ownerHex models.HexID, name, domain string) (*models.Domain, error)
	// CreateModerator creates and persists a new domain moderator record
	CreateModerator(domain, email string) (*models.DomainModerator, error)
	// CreateSSOSecret generates a new SSO secret token for the given domain and saves that in the domain properties
	CreateSSOSecret(domain string) (models.HexID, error)
	// CreateSSOToken generates, persists, and returns a new SSO token for the given domain and commenter token
	CreateSSOToken(domain string, commenterToken models.CommenterHexID) (models.HexID, error)
	// Delete deletes the specified domain
	Delete(domain string) error
	// DeleteModerator deletes the specified domain moderator
	DeleteModerator(domain, email string) error
	// FindByName fetches and returns a domain with the specified name
	FindByName(domainName string) (*models.Domain, error)
	// IsDomainModerator returns whether the given email is a moderator in the given domain
	IsDomainModerator(email, domain string) (bool, error)
	// IsDomainOwner returns whether the given owner hex ID is an owner of the given domain
	IsDomainOwner(id models.HexID, domain string) (bool, error)
	// ListByOwner fetches and returns a list of domains for the specified owner
	ListByOwner(ownerHex models.HexID) ([]*models.Domain, error)
	// RegisterView records a domain view in the database. commenterHex should be "anonymous" for an unauthenticated
	// viewer
	RegisterView(domain string, commenterHex models.CommenterHexID) error
	// StatsForComments collects and returns comment statistics for the given domain
	StatsForComments(domain string) ([]int64, error)
	// StatsForViews collects and returns view statistics for the given domain
	StatsForViews(domain string) ([]int64, error)
	// TakeSSOToken queries and removes the provided token from the database, returning its domain and commenter token
	TakeSSOToken(token models.HexID) (string, models.CommenterHexID, error)
	// Update updates the domain record in the database
	Update(domain *models.Domain) error
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

func (svc *domainService) Create(ownerHex models.HexID, name, domain string) (*models.Domain, error) {
	logger.Debugf("domainService.Create(%s, %s, %s)", ownerHex, name, domain)

	// Insert a new record
	d := models.Domain{
		CreationDate: strfmt.DateTime(time.Now().UTC()),
		Domain:       domain,
		Name:         name,
		OwnerHex:     ownerHex,
	}
	err := db.Exec(
		"insert into domains(ownerhex, name, domain, creationdate) values($1, $2, $3, $4);",
		d.OwnerHex, d.Name, d.Domain, d.CreationDate)
	if err != nil {
		logger.Errorf("domainService.Create: Exec() failed: %v", err)
		return nil, translateDBErrors(err)
	}

	// Succeeded
	return &d, nil
}

func (svc *domainService) CreateModerator(domain, email string) (*models.DomainModerator, error) {
	logger.Debugf("domainService.CreateModerator(%s, %s)", domain, email)

	// Create a new email record
	if _, err := TheEmailService.Create(email); err != nil {
		return nil, err
	}

	// Create a new domain moderator record
	dm := models.DomainModerator{
		AddDate: strfmt.DateTime(time.Now().UTC()),
		Domain:  domain,
		Email:   strfmt.Email(email),
	}
	err := db.Exec("insert into moderators(domain, email, adddate) values($1, $2, $3);", dm.Domain, dm.Email, dm.AddDate)
	if err != nil {
		logger.Errorf("domainService.CreateModerator: Exec() failed: %v", err)
		return nil, translateDBErrors(err)
	}

	// Succeeded
	return &dm, nil
}

func (svc *domainService) CreateSSOSecret(domain string) (models.HexID, error) {
	logger.Debugf("domainService.CreateSSOSecret(%s)", domain)

	// Generate a new token
	token, err := data.RandomHexID()
	if err != nil {
		logger.Errorf("userService.CreateSSOSecret: RandomHexID() failed: %v", err)
		return "", err
	}

	// Update the domain record
	if err = db.Exec("update domains set ssosecret=$1 where domain=$2;", token, domain); err != nil {
		logger.Errorf("domainService.CreateSSOSecret: Exec() failed: %v", err)
		return "", translateDBErrors(err)
	}

	// Succeeded
	return token, nil
}

func (svc *domainService) CreateSSOToken(domain string, commenterToken models.CommenterHexID) (models.HexID, error) {
	logger.Debugf("domainService.CreateSSOToken(%s, %s)", domain, commenterToken)

	// Generate a new token
	token, err := data.RandomHexID()
	if err != nil {
		logger.Errorf("userService.CreateSSOToken: RandomHexID() failed: %v", err)
		return "", err
	}

	// Insert a new token record
	err = db.Exec(
		"insert into ssotokens(token, domain, commentertoken, creationdate) values($1, $2, $3, $4);",
		token, domain, commenterToken, time.Now().UTC())
	if err != nil {
		logger.Errorf("domainService.CreateSSOToken: Exec() failed: %v", err)
		return "", translateDBErrors(err)
	}

	// Succeeded
	return token, nil
}

func (svc *domainService) Delete(domain string) error {
	logger.Debugf("domainService.Delete(%s)", domain)

	// Remove the domain's view stats, moderators, ssotokens
	err := checkErrors(
		db.Exec(
			"delete from views where domain=$1;"+
				"delete from moderators where domain=$1;"+
				"delete from ssotokens where domain=$1;",
			domain))
	if err != nil {
		logger.Errorf("domainService.Delete: Exec() failed for dependent object: %v", err)
		return translateDBErrors(err)
	}

	// Remove the domain itself
	if err := db.Exec("delete from domains where domain=$1;", domain); err != nil {
		logger.Errorf("domainService.Delete: Exec() failed for domain: %v", err)
		return translateDBErrors(err)
	}

	// Succeeded
	return nil
}

func (svc *domainService) DeleteModerator(domain, email string) error {
	logger.Debugf("domainService.DeleteModerator(%s, %s)", domain, email)

	// Remove the row from the database
	if err := db.Exec("delete from moderators where domain=$1 and email=$2;", domain, email); err != nil {
		logger.Errorf("domainService.DeleteModerator: Exec() failed: %v", err)
		return translateDBErrors(err)
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
		return nil, translateDBErrors(err)
	}
	defer rows.Close()

	// Fetch the domain(s)
	if domains, err := svc.fetchDomainsAndModerators(rows); err != nil {
		return nil, translateDBErrors(err)
	} else if len(domains) == 0 {
		return nil, ErrNotFound
	} else {
		// Grab the first one
		return domains[0], nil
	}
}

func (svc *domainService) IsDomainModerator(email, domain string) (bool, error) {
	logger.Debugf("domainService.IsDomainModerator(%s, %s)", email, domain)

	// Query the row
	row := db.QueryRow("select 1 from moderators where domain=$1 and email=$2;", domain, email)
	var b byte
	if err := row.Scan(&b); err == sql.ErrNoRows {
		// No rows means it isn't a moderator
		return false, nil

	} else if err != nil {
		// Any other database error
		logger.Errorf("domainService.IsDomainModerator: QueryRow() failed: %v", err)
		return false, translateDBErrors(err)
	}

	// Succeeded: the email belongs to a domain moderator
	return true, nil
}

func (svc *domainService) IsDomainOwner(id models.HexID, domain string) (bool, error) {
	logger.Debugf("domainService.IsDomainOwner(%s, %s)", id, domain)

	// Query the row
	row := db.QueryRow("select 1 from domains where ownerhex=$1 and domain=$2", id, domain)
	var b byte
	if err := row.Scan(&b); err == sql.ErrNoRows {
		// No rows means it isn't an owner
		return false, nil

	} else if err != nil {
		// Any other database error
		logger.Errorf("domainService.IsDomainOwner: QueryRow() failed: %v", err)
		return false, translateDBErrors(err)
	}

	// Succeeded: the ID belongs to a domain owner
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
		return nil, translateDBErrors(err)
	}
	defer rows.Close()

	// Fetch the domains
	if domains, err := svc.fetchDomainsAndModerators(rows); err != nil {
		return nil, translateDBErrors(err)
	} else {
		return domains, nil
	}
}

func (svc *domainService) RegisterView(domain string, commenterHex models.CommenterHexID) error {
	logger.Debugf("domainService.RegisterView(%s, %s)", domain, commenterHex)

	// Insert a new view record
	err := db.Exec(
		"insert into views(domain, commenterhex, viewdate) values ($1, $2, $3);",
		domain, commenterHex, time.Now().UTC())
	if err != nil {
		logger.Warningf("domainService.RegisterView: Exec() failed: %v", err)
		return translateDBErrors(err)
	}

	// Succeeded
	return nil
}

func (svc *domainService) StatsForComments(domain string) ([]int64, error) {
	logger.Debugf("domainService.StatsForComments(%s)", domain)

	// Query the data from the database, grouped by day
	rows, err := db.Query(
		"select count(c.creationdate) "+
			"from (select to_char(date_trunc('day', (current_date-offs)), 'YYYY-MM-DD') as date from generate_series(0, 30, 1) as offs) d "+
			"left join comments c on d.date=to_char(date_trunc('day', c.creationdate), 'YYYY-MM-DD') and c.domain=$1 "+
			"group by d.date "+
			"order by d.date;",
		domain)
	if err != nil {
		logger.Errorf("domainService.StatsForComments: Query() failed: %v", err)
		return nil, translateDBErrors(err)
	}
	defer rows.Close()

	// Collect the data
	if res, err := svc.fetchStats(rows); err != nil {
		return nil, translateDBErrors(err)
	} else {
		// Succeeded
		return res, nil
	}
}

func (svc *domainService) StatsForViews(domain string) ([]int64, error) {
	logger.Debugf("domainService.StatsForViews(%s)", domain)

	// Query the data from the database, grouped by day
	rows, err := db.Query(
		"select count(v.viewdate) "+
			"from (select to_char(date_trunc('day', (current_date-offs)), 'YYYY-MM-DD') as date from generate_series(0, 30, 1) as offs) d "+
			"left join views v on d.date = to_char(date_trunc('day', v.viewdate), 'YYYY-MM-DD') and v.domain=$1 "+
			"group by d.date "+
			"order by d.date;",
		domain)
	if err != nil {
		logger.Errorf("domainService.StatsForViews: Query() failed: %v", err)
		return nil, translateDBErrors(err)
	}
	defer rows.Close()

	// Collect the data
	if res, err := svc.fetchStats(rows); err != nil {
		return nil, translateDBErrors(err)
	} else {
		// Succeeded
		return res, nil
	}
}

func (svc *domainService) TakeSSOToken(token models.HexID) (string, models.CommenterHexID, error) {
	logger.Debugf("domainService.TakeSSOToken(%s)", token)

	// Fetch and delete the token
	row := db.QueryRow("delete from ssotokens where token=$1 returning domain, commentertoken;", token)
	var domain string
	var commenterToken models.CommenterHexID
	if err := row.Scan(&domain, &commenterToken); err != nil {
		logger.Errorf("domainService.TakeSSOToken: Scan() failed: %v", err)
		return "", "", translateDBErrors(err)
	}

	// Succeeded
	return domain, commenterToken, nil
}

func (svc *domainService) Update(domain *models.Domain) error {
	logger.Debug("domainService.Update(...)")

	// Update the domain
	err := db.Exec(
		"update domains "+
			"set name=$1, state=$2, autospamfilter=$3, requiremoderation=$4, requireidentification=$5, "+
			"moderateallanonymous=$6, emailnotificationpolicy=$7, commentoprovider=$8, googleprovider=$9, "+
			"githubprovider=$10, gitlabprovider=$11, twitterprovider=$12, ssoprovider=$13, ssourl=$14, "+
			"defaultsortpolicy=$15 "+
			"where domain=$16;",
		domain.Name,
		domain.State,
		domain.AutoSpamFilter,
		domain.RequireModeration,
		domain.RequireIdentification,
		domain.ModerateAllAnonymous,
		domain.EmailNotificationPolicy,
		domain.Idps["commento"],
		domain.Idps["google"],
		domain.Idps["github"],
		domain.Idps["gitlab"],
		domain.Idps["twitter"],
		domain.Idps["sso"],
		domain.SsoURL,
		domain.DefaultSortPolicy,
		domain.Domain)
	if err != nil {
		logger.Errorf("domainService.Update: Exec() failed: %v", err)
		return translateDBErrors(err)
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

// fetchStats collects and returns a daily statistics using the provided database rows
func (svc *domainService) fetchStats(rs *sql.Rows) ([]int64, error) {
	// Collect the data
	var res []int64
	for rs.Next() {
		var i int64
		if err := rs.Scan(&i); err != nil {
			logger.Errorf("domainService.fetchStats: rs.Scan() failed: %v", err)
			return nil, err
		}
		res = append(res, i)
	}

	// Check that Next() didn't error
	if err := rs.Err(); err != nil {
		return nil, err
	}

	// Succeeded
	return res, nil
}
