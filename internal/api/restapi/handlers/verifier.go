package handlers

import (
	"github.com/go-openapi/runtime/middleware"
	"gitlab.com/comentario/comentario/internal/api/models"
	"gitlab.com/comentario/comentario/internal/data"
	"gitlab.com/comentario/comentario/internal/svc"
	"gitlab.com/comentario/comentario/internal/util"
)

// Verifier is a global VerifierService implementation
var Verifier VerifierService = &verifier{}

// VerifierService is an API service interface for data and permission verification
type VerifierService interface {
	// CommenterLocalEmaiUnique verifies there's no existing commenter  user using the password authentication with the
	// given email
	CommenterLocalEmaiUnique(email string) middleware.Responder
	// OwnerEmaiUnique verifies there's no existing owner user with the given email
	OwnerEmaiUnique(email string) middleware.Responder
	// PrincipalIsAuthenticated verifies the given principal is an authenticated one
	PrincipalIsAuthenticated(principal data.Principal) middleware.Responder
	// UserIsDomainModerator verifies the owner with the given email is a moderator in the specified domain
	UserIsDomainModerator(email, domainName string) middleware.Responder
	// UserOwnsDomain verifies the owner with the given hex ID owns the specified domain
	UserOwnsDomain(id models.HexID, domainName string) middleware.Responder
}

// ----------------------------------------------------------------------------------------------------------------------
// verifier is a blueprint VerifierService implementation
type verifier struct{}

func (v *verifier) CommenterLocalEmaiUnique(email string) middleware.Responder {
	// Verify no such email is registered yet
	if _, err := svc.TheUserService.FindCommenterByIdPEmail("", email, false); err == nil {
		return respBadRequest(util.ErrorEmailAlreadyExists)
	} else if err != svc.ErrNotFound {
		return respServiceError(err)
	}
	return nil
}

func (v *verifier) OwnerEmaiUnique(email string) middleware.Responder {
	// Verify no such email is registered yet
	if _, err := svc.TheUserService.FindOwnerByEmail(email, false); err == nil {
		return respBadRequest(util.ErrorEmailAlreadyExists)
	} else if err != svc.ErrNotFound {
		return respServiceError(err)
	}
	return nil
}

func (v *verifier) PrincipalIsAuthenticated(principal data.Principal) middleware.Responder {
	if principal.IsAnonymous() {
		return respUnauthorized(util.ErrorUnauthenticated)
	}
	return nil
}

func (v *verifier) UserIsDomainModerator(email, domainName string) middleware.Responder {
	if b, err := svc.TheDomainService.IsDomainModerator(email, domainName); err != nil {
		return respServiceError(err)
	} else if !b {
		return respForbidden(util.ErrorNotModerator)
	}
	return nil
}

func (v *verifier) UserOwnsDomain(id models.HexID, domainName string) middleware.Responder {
	if b, err := svc.TheDomainService.IsDomainOwner(id, domainName); err != nil {
		return respServiceError(err)
	} else if !b {
		return respForbidden(util.ErrorNotDomainOwner)
	}
	return nil
}
