package svc

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
