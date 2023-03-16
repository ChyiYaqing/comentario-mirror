package handlers

import (
	"bytes"
	"github.com/disintegration/imaging"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/swag"
	"gitlab.com/comentario/comentario/internal/api/models"
	"gitlab.com/comentario/comentario/internal/api/restapi/operations"
	"gitlab.com/comentario/comentario/internal/config"
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
)

func CommenterLogin(params operations.CommenterLoginParams) middleware.Responder {
	// Try to find a local user with the given email
	commenter, err := svc.TheUserService.FindCommenterByIdPEmail("", data.EmailToString(params.Body.Email), true)
	if err != nil {
		return respUnauthorized(util.ErrorInvalidEmailPassword)
	}

	// Verify the provided password
	if err := bcrypt.CompareHashAndPassword([]byte(commenter.PasswordHash), []byte(swag.StringValue(params.Body.Password))); err != nil {
		return respUnauthorized(util.ErrorInvalidEmailPassword)
	}

	// Create a new commenter session
	commenterToken, err := svc.TheUserService.CreateCommenterSession(commenter.CommenterHexID())
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
		Success:        true,
	})
}

func CommenterNew(params operations.CommenterNewParams) middleware.Responder {
	email := data.EmailToString(params.Body.Email)
	name := data.TrimmedString(params.Body.Name)
	website := strings.TrimSpace(params.Body.Website)

	// TODO this is awful
	if website == "" {
		website = "undefined"
	}

	if _, err := svc.TheUserService.CreateCommenter(email, name, website, "undefined", "", *params.Body.Password); err != nil {
		return respServiceError(err)
	}

	// Succeeded
	return operations.NewCommenterNewOK().WithPayload(&operations.CommenterNewOKBody{
		ConfirmEmail: config.SMTPConfigured,
		Success:      true,
	})
}

func CommenterPhoto(params operations.CommenterPhotoParams) middleware.Responder {
	// Find the commenter user
	commenter, err := svc.TheUserService.FindCommenterByID(models.CommenterHexID(params.CommenterHex))
	if err != nil {
		return respServiceError(err)
	}

	// Fetch the image pointed to by the PhotoURL
	resp, err := http.Get(commenter.PhotoURL)
	if err != nil {
		return operations.NewGenericNotFound()
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
	// Find the commenter
	commenter, err := svc.TheUserService.FindCommenterByToken(*params.Body.CommenterToken)
	if err == svc.ErrNotFound {
		// Not logged in or session doesn't exist
		return operations.NewCommenterSelfOK()

	} else if err != nil {
		// Any other error
		return respServiceError(err)
	}

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

func CommenterTokenNew(operations.CommenterTokenNewParams) middleware.Responder {
	// Create an "anonymous" session
	token, err := svc.TheUserService.CreateCommenterSession("")
	if err != nil {
		return respServiceError(err)
	}

	// Succeeded
	return operations.NewCommenterTokenNewOK().WithPayload(&operations.CommenterTokenNewOKBody{
		CommenterToken: token,
		Success:        true,
	})
}

func CommenterUpdate(params operations.CommenterUpdateParams) middleware.Responder {
	// Find the commenter
	commenter, err := svc.TheUserService.FindCommenterByToken(*params.Body.CommenterToken)
	if err != nil {
		return respServiceError(err)
	}

	// Only locally authenticated users can be updated
	if commenter.Provider != "commento" {
		return respBadRequest(util.ErrorCannotUpdateOauthProfile)
	}

	// Update the commenter in the database
	err = svc.TheUserService.UpdateCommenter(
		commenter.CommenterHexID(),
		commenter.Email,
		data.TrimmedString(params.Body.Name),
		strings.TrimSpace(params.Body.Link),
		strings.TrimSpace(params.Body.Photo),
		commenter.Provider)
	if err != nil {
		return respServiceError(err)
	}

	// Succeeded
	return operations.NewCommenterUpdateOK().WithPayload(&models.APIResponseBase{Success: true})
}
