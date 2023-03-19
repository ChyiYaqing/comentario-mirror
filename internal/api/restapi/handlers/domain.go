package handlers

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/markbates/goth"
	"gitlab.com/comentario/comentario/internal/api/exmodels"
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

	// Verify the user owns the domain
	domain := data.TrimmedString(params.Body.Domain)
	if r := Verifier.UserOwnsDomain(user.HexID, domain); r != nil {
		return r
	}

	// Clear all domain's pages/comments/votes
	if err = svc.TheDomainService.Clear(domain); err != nil {
		return respServiceError(err)
	}

	// Succeeded
	return operations.NewDomainClearNoContent()
}

func DomainDelete(params operations.DomainDeleteParams) middleware.Responder {
	user, err := svc.TheUserService.FindOwnerByToken(*params.Body.OwnerToken)
	if err != nil {
		return respServiceError(err)
	}

	// Verify the user owns the domain
	domain := data.TrimmedString(params.Body.Domain)
	if r := Verifier.UserOwnsDomain(user.HexID, domain); r != nil {
		return r
	}

	// Delete the domain
	if err = svc.TheDomainService.Delete(domain); err != nil {
		return respServiceError(err)
	}

	// Succeeded
	return operations.NewDomainDeleteNoContent()
}

func DomainList(_ operations.DomainListParams, principal data.Principal) middleware.Responder {
	// Fetch domains by the owner
	domains, err := svc.TheDomainService.ListByOwner(principal.GetUser().HexID)
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
	})
}

func DomainModeratorDelete(params operations.DomainModeratorDeleteParams) middleware.Responder {
	user, err := svc.TheUserService.FindOwnerByToken(*params.Body.OwnerToken)
	if err != nil {
		return respServiceError(err)
	}

	// Verify the user owns the domain
	domain := data.TrimmedString(params.Body.Domain)
	if r := Verifier.UserOwnsDomain(user.HexID, domain); r != nil {
		return r
	}

	// Delete the moderator from the database
	if err = svc.TheDomainService.DeleteModerator(domain, data.EmailToString(params.Body.Email)); err != nil {
		return respServiceError(err)
	}

	// Succeeded
	return operations.NewDomainModeratorDeleteNoContent()
}

func DomainModeratorNew(params operations.DomainModeratorNewParams) middleware.Responder {
	user, err := svc.TheUserService.FindOwnerByToken(*params.Body.OwnerToken)
	if err != nil {
		return respServiceError(err)
	}

	// Verify the user owns the domain
	domain := data.TrimmedString(params.Body.Domain)
	if r := Verifier.UserOwnsDomain(user.HexID, domain); r != nil {
		return r
	}

	// Register a new domain moderator
	if _, err := svc.TheDomainService.CreateModerator(domain, data.EmailToString(params.Body.Email)); err != nil {
		return respServiceError(err)
	}

	// Succeeded
	return operations.NewDomainModeratorNewNoContent()
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
	return operations.NewDomainNewOK().WithPayload(&operations.DomainNewOKBody{Domain: domain.Domain})
}

func DomainSsoSecretNew(params operations.DomainSsoSecretNewParams) middleware.Responder {
	user, err := svc.TheUserService.FindOwnerByToken(*params.Body.OwnerToken)
	if err != nil {
		return respServiceError(err)
	}

	// Verify the user owns the domain
	domain := data.TrimmedString(params.Body.Domain)
	if r := Verifier.UserOwnsDomain(user.HexID, domain); r != nil {
		return r
	}

	// Generate a new SSO secret for the domain
	token, err := svc.TheDomainService.CreateSSOSecret(domain)
	if err != nil {
		return respServiceError(err)
	}

	// Succeeded
	return operations.NewDomainSsoSecretNewOK().WithPayload(&operations.DomainSsoSecretNewOKBody{SsoSecret: token})
}

func DomainStatistics(params operations.DomainStatisticsParams) middleware.Responder {
	user, err := svc.TheUserService.FindOwnerByToken(*params.Body.OwnerToken)
	if err != nil {
		return respServiceError(err)
	}

	// Verify the user owns the domain
	domain := data.TrimmedString(params.Body.Domain)
	if r := Verifier.UserOwnsDomain(user.HexID, domain); r != nil {
		return r
	}

	// Collect view stats
	views, err := svc.TheDomainService.StatsForViews(domain)
	if err != nil {
		return respServiceError(err)
	}

	// Collect comment stats
	comments, err := svc.TheDomainService.StatsForComments(domain)
	if err != nil {
		return respServiceError(err)
	}

	// Succeeded
	return operations.NewDomainStatisticsOK().WithPayload(&operations.DomainStatisticsOKBody{
		CommentsLast30Days: comments,
		ViewsLast30Days:    views,
	})
}

func DomainUpdate(params operations.DomainUpdateParams) middleware.Responder {
	// Find the owner user
	user, err := svc.TheUserService.FindOwnerByToken(*params.Body.OwnerToken)
	if err != nil {
		return respServiceError(err)
	}

	// Verify the user owns the domain
	domain := params.Body.Domain
	if r := Verifier.UserOwnsDomain(user.HexID, domain.Domain); r != nil {
		return r
	}

	// Validate SSO provider
	if domain.Idps["sso"] && domain.SsoURL == "" {
		return respBadRequest(util.ErrorSSOURLMissing)
	}

	// Update the domain record
	if err := svc.TheDomainService.Update(domain); err != nil {
		return respServiceError(err)
	}

	// Succeeded
	return operations.NewDomainUpdateNoContent()
}
