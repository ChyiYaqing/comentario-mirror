package handlers

import (
	"bytes"
	"github.com/disintegration/imaging"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/swag"
	"gitlab.com/comentario/comentario/internal/api/models"
	"gitlab.com/comentario/comentario/internal/api/restapi/operations"
	"gitlab.com/comentario/comentario/internal/data"
	"gitlab.com/comentario/comentario/internal/svc"
	"gitlab.com/comentario/comentario/internal/util"
	"golang.org/x/crypto/bcrypt"
	"image"
	"image/color"
	"image/draw"
	"io"
	"net/http"
	"strings"
	"time"
)

func CommenterLogin(params operations.CommenterLoginParams) middleware.Responder {
	// Try to find a local user with the given email
	commenter, err := svc.TheUserService.FindCommenterByIdPEmail("", data.EmailToString(params.Body.Email), true)
	if err != nil {
		time.Sleep(util.WrongAuthDelay)
		return respUnauthorized(util.ErrorInvalidEmailPassword)
	}

	// Verify the provided password
	if err := bcrypt.CompareHashAndPassword([]byte(commenter.PasswordHash), []byte(swag.StringValue(params.Body.Password))); err != nil {
		time.Sleep(util.WrongAuthDelay)
		return respUnauthorized(util.ErrorInvalidEmailPassword)
	}

	// Create a new commenter session
	commenterToken, err := svc.TheUserService.CreateCommenterSession(commenter.HexID)
	if err != nil {
		return respServiceError(err)
	}

	// Fetch the commenter's email
	email, err := svc.TheEmailService.FindByEmail(commenter.Email)
	if err != nil {
		return respServiceError(err)
	}

	// Succeeded
	return operations.NewCommenterLoginOK().WithPayload(&operations.CommenterLoginOKBody{
		Commenter:      commenter.ToCommenter(),
		CommenterToken: commenterToken,
		Email:          email,
	})
}

func CommenterLogout(params operations.CommenterLogoutParams, principal data.Principal) middleware.Responder {
	// Verify the commenter is authenticated
	if r := Verifier.PrincipalIsAuthenticated(principal); r != nil {
		return r
	}

	// Extract a commenter token from the corresponding header, if any
	if token := models.HexID(params.HTTPRequest.Header.Get(util.HeaderCommenterToken)); token.Validate(nil) == nil {
		// Delete the commenter token, ignoring any error
		_ = svc.TheUserService.DeleteCommenterSession(principal.GetHexID(), token)
	}

	// Regardless of whether the above was successful, return a success response
	return operations.NewCommenterLogoutNoContent()
}

func CommenterNew(params operations.CommenterNewParams) middleware.Responder {
	email := data.EmailToString(params.Body.Email)
	name := data.TrimmedString(params.Body.Name)
	website := strings.TrimSpace(params.Body.Website)

	// Since the local authentication is used, verify the email is unique
	if r := Verifier.CommenterLocalEmaiUnique(email); r != nil {
		return r
	}

	// Create a commenter record in the database
	if _, err := svc.TheUserService.CreateCommenter(email, name, website, "", "", *params.Body.Password); err != nil {
		return respServiceError(err)
	}

	// Succeeded
	return operations.NewCommenterNewNoContent()
}

func CommenterPhoto(params operations.CommenterPhotoParams) middleware.Responder {
	// Validate the passed commenter hex ID
	id := models.HexID(params.CommenterHex)
	if err := id.Validate(nil); err != nil {
		return respBadRequest(err)
	}

	// Find the commenter user
	commenter, err := svc.TheUserService.FindCommenterByID(id)
	if err != nil {
		return respServiceError(err)
	}

	// Fetch the image pointed to by the PhotoURL
	resp, err := http.Get(commenter.PhotoURL)
	if err != nil {
		return respNotFound()
	}
	defer resp.Body.Close()

	// Limit the size of the response to 512 KiB to prevent DoS attacks that exhaust memory
	limitedResp := &io.LimitedReader{R: resp.Body, N: 512 * 1024}

	// Decode the image
	img, imgFormat, err := image.Decode(limitedResp)
	if err != nil {
		return operations.NewGenericInternalServerError().
			WithPayload(&operations.GenericInternalServerErrorBody{Details: "Failed to decode image"})
	}
	logger.Debugf("Loaded commenter avatar: format=%s, dimensions=%s", imgFormat, img.Bounds().Size().String())

	// If it's a PNG, flatten it against a white background
	if imgFormat == "png" {
		logger.Debug("Flattening PNG image")

		// Create a new white Image with the same dimension of PNG image
		bgImage := image.NewRGBA(img.Bounds())
		draw.Draw(bgImage, bgImage.Bounds(), &image.Uniform{C: color.White}, image.Point{}, draw.Src)

		// Paste the PNG image over the background
		draw.Draw(bgImage, bgImage.Bounds(), img, img.Bounds().Min, draw.Over)
		img = bgImage
	}

	// Resize the image and encode into a JPEG
	var buf bytes.Buffer
	if err = imaging.Encode(&buf, imaging.Resize(img, 38, 0, imaging.Lanczos), imaging.JPEG); err != nil {
		return operations.NewGenericInternalServerError().
			WithPayload(&operations.GenericInternalServerErrorBody{Details: "Failed to encode image"})
	}
	return operations.NewCommenterPhotoOK().WithPayload(io.NopCloser(&buf))
}

func CommenterSelf(params operations.CommenterSelfParams) middleware.Responder {
	// Extract a commenter token from the corresponding header, if any
	if token := models.HexID(params.HTTPRequest.Header.Get(util.HeaderCommenterToken)); token.Validate(nil) == nil {
		// Find the commenter
		if commenter, err := svc.TheUserService.FindCommenterByToken(token); err != nil && err != svc.ErrNotFound {
			// Any error except "not found"
			return respServiceError(err)

		} else if err == nil {
			// Fetch the commenter's email
			email, err := svc.TheEmailService.FindByEmail(commenter.Email)
			if err != nil {
				return respServiceError(err)
			}

			// Succeeded
			return operations.NewCommenterSelfOK().WithPayload(&operations.CommenterSelfOKBody{
				Commenter: commenter.ToCommenter(),
				Email:     email,
			})
		}
	}

	// Not logged in, bad token, commenter is anonymous or doesn't exist
	return operations.NewCommenterSelfNoContent()
}

func CommenterTokenNew(operations.CommenterTokenNewParams) middleware.Responder {
	// Create an "anonymous" session
	token, err := svc.TheUserService.CreateCommenterSession("")
	if err != nil {
		return respServiceError(err)
	}

	// Succeeded
	return operations.NewCommenterTokenNewOK().WithPayload(&operations.CommenterTokenNewOKBody{CommenterToken: token})
}

func CommenterUpdate(params operations.CommenterUpdateParams, principal data.Principal) middleware.Responder {
	// Verify the commenter is authenticated
	if r := Verifier.PrincipalIsAuthenticated(principal); r != nil {
		return r
	}
	commenter := principal.(*data.UserCommenter)

	// Only locally authenticated users can be updated
	if commenter.Provider != "" {
		return respBadRequest(util.ErrorCannotUpdateOauthProfile)
	}

	// Update the commenter in the database
	err := svc.TheUserService.UpdateCommenter(
		commenter.HexID,
		commenter.Email,
		data.TrimmedString(params.Body.Name),
		strings.TrimSpace(params.Body.Link),
		strings.TrimSpace(params.Body.Photo),
		commenter.Provider)
	if err != nil {
		return respServiceError(err)
	}

	// Succeeded
	return operations.NewCommenterUpdateNoContent()
}
