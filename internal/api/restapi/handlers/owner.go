package handlers

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/swag"
	"gitlab.com/comentario/comentario/internal/api/models"
	"gitlab.com/comentario/comentario/internal/api/restapi/operations"
	"gitlab.com/comentario/comentario/internal/config"
	"gitlab.com/comentario/comentario/internal/data"
	"gitlab.com/comentario/comentario/internal/svc"
	"gitlab.com/comentario/comentario/internal/util"
	"golang.org/x/crypto/bcrypt"
	"time"
)

func OwnerConfirmHex(params operations.OwnerConfirmHexParams) middleware.Responder {
	// Update the owner, if the token checks out
	conf := "true"
	if err := svc.TheUserService.ConfirmOwner(models.HexID(params.ConfirmHex)); err != nil {
		conf = "false"
	}

	// Redirect to login
	return operations.NewOwnerConfirmHexTemporaryRedirect().
		WithLocation(config.URLFor("login", map[string]string{"confirmed": conf}))
}

func OwnerDelete(params operations.OwnerDeleteParams) middleware.Responder {
	// Find the owner user
	user, err := svc.TheUserService.FindOwnerByToken(*params.Body.OwnerToken)
	if err != nil {
		return respServiceError(err)
	}

	// Fetch a list of domains
	if domains, err := svc.TheDomainService.ListByOwner(user.HexID); err != nil {
		return respServiceError(err)

		// Make sure the owner owns no domains
	} else if len(domains) > 0 {
		return respBadRequest(util.ErrorCannotDeleteOwner)
	}

	// Remove the owner user
	if err := svc.TheUserService.DeleteOwnerByID(user.HexID); err != nil {
		return respServiceError(err)
	}

	// Succeeded
	return operations.NewOwnerDeleteNoContent()
}

func OwnerLogin(params operations.OwnerLoginParams) middleware.Responder {
	// Find the owner
	owner, err := svc.TheUserService.FindOwnerByEmail(data.EmailToString(params.Body.Email), true)
	if err == svc.ErrNotFound {
		time.Sleep(util.WrongAuthDelay)
		return respUnauthorized(util.ErrorInvalidEmailPassword)
	} else if err != nil {
		return respServiceError(err)
	}

	// Verify the owner is confirmed
	if !owner.EmailConfirmed {
		return respUnauthorized(util.ErrorUnconfirmedEmail)
	}

	// Verify the provided password
	if err := bcrypt.CompareHashAndPassword([]byte(owner.PasswordHash), []byte(swag.StringValue(params.Body.Password))); err != nil {
		time.Sleep(util.WrongAuthDelay)
		return respUnauthorized(util.ErrorInvalidEmailPassword)
	}

	// Create a new owner session
	ownerToken, err := svc.TheUserService.CreateOwnerSession(owner.HexID)
	if err != nil {
		return respServiceError(err)
	}

	// Succeeded
	return operations.NewOwnerLoginOK().WithPayload(&operations.OwnerLoginOKBody{OwnerToken: ownerToken})
}

func OwnerNew(params operations.OwnerNewParams) middleware.Responder {
	// Verify new owners are allowed
	if !config.CLIFlags.AllowNewOwners {
		return respForbidden(util.ErrorNewOwnerForbidden)
	}

	// Verify no owner with that email exists yet
	email := data.EmailToString(params.Body.Email)
	if r := Verifier.OwnerEmaiUnique(email); r != nil {
		return r
	}

	// Create a new email record
	if _, err := svc.TheEmailService.Create(email); err != nil {
		return respServiceError(err)
	}

	// Create a new owner record
	name := data.TrimmedString(params.Body.Name)
	pwd := swag.StringValue(params.Body.Password)
	owner, err := svc.TheUserService.CreateOwner(email, name, pwd)
	if err != nil {
		return respServiceError(err)
	}

	// If mailing is configured, create and mail a confirmation token
	if config.SMTPConfigured {
		// Create a new confirmation token
		token, err := svc.TheUserService.CreateOwnerConfirmationToken(owner.HexID)
		if err != nil {
			return respServiceError(err)
		}

		// Send a confirmation email
		err = svc.TheMailService.SendFromTemplate(
			"",
			email,
			"Please confirm your email address",
			"confirm-hex.gohtml",
			map[string]any{"URL": config.URLForAPI("owner/confirm-hex", map[string]string{"confirmHex": string(token)})})
		if err != nil {
			return respServiceError(err)
		}
	}

	// If no commenter with that email exists yet, register the owner also as a commenter, with the same password
	if _, err := svc.TheUserService.FindCommenterByIdPEmail("", email, false); err == svc.ErrNotFound {
		if _, err := svc.TheUserService.CreateCommenter(email, name, "", "", "", pwd); err != nil {
			return respServiceError(err)
		}
	}

	// Succeeded
	return operations.NewOwnerNewOK().WithPayload(&operations.OwnerNewOKBody{ConfirmEmail: config.SMTPConfigured})
}

func OwnerSelf(params operations.OwnerSelfParams) middleware.Responder {
	// Try to find the owner
	user, err := svc.TheUserService.FindOwnerByToken(*params.Body.OwnerToken)
	if err == svc.ErrNotFound {
		// Owner isn't logged id
		return operations.NewOwnerSelfNoContent()
	} else if err != nil {
		// Any other database error
		return respServiceError(err)
	}

	// Succeeded: owner's logged in
	return operations.NewOwnerSelfOK().WithPayload(&operations.OwnerSelfOKBody{Owner: user.ToOwner()})
}
