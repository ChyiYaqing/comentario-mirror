package handlers

import (
	"bytes"
	"fmt"
	"github.com/disintegration/imaging"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"gitlab.com/comentario/comentario/internal/api/models"
	"gitlab.com/comentario/comentario/internal/api/restapi/operations"
	"gitlab.com/comentario/comentario/internal/config"
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

const commentersRowColumns = `
	commenters.commenterHex,
	commenters.email,
	commenters.name,
	commenters.link,
	commenters.photo,
	commenters.provider,
	commenters.joinDate
`

func CommenterLogin(params operations.CommenterLoginParams) middleware.Responder {
	commenterToken, err := commenterLogin(*params.Body.Email, *params.Body.Password)
	if err != nil {
		return operations.NewCommenterLoginOK().WithPayload(&operations.CommenterLoginOKBody{Message: err.Error()})
	}

	// TODO: modify commenterLogin to directly return c?
	commenter, err := commenterGetByCommenterToken(commenterToken)
	if err != nil {
		return operations.NewCommenterLoginOK().WithPayload(&operations.CommenterLoginOKBody{Message: err.Error()})
	}

	email, err := emailGet(commenter.Email)
	if err != nil {
		return operations.NewCommenterLoginOK().WithPayload(&operations.CommenterLoginOKBody{Message: err.Error()})
	}

	// Succeeded
	return operations.NewCommenterLoginOK().WithPayload(&operations.CommenterLoginOKBody{
		Commenter:      commenter,
		CommenterToken: commenterToken,
		Email:          email,
		Success:        true,
	})
}

func CommenterNew(params operations.CommenterNewParams) middleware.Responder {
	website := strings.TrimSpace(params.Body.Website)

	// TODO this is awful
	if website == "" {
		website = "undefined"
	}

	if _, err := commenterNew(*params.Body.Email, *params.Body.Name, website, "undefined", "commento", *params.Body.Password); err != nil {
		return operations.NewCommenterNewOK().WithPayload(&operations.CommenterNewOKBody{Message: err.Error()})
	}

	// Succeeded
	return operations.NewCommenterNewOK().WithPayload(&operations.CommenterNewOKBody{
		ConfirmEmail: config.SMTPConfigured,
		Success:      true,
	})
}

func CommenterPhoto(params operations.CommenterPhotoParams) middleware.Responder {
	c, err := commenterGetByHex(models.CommenterHexID(params.CommenterHex))
	if err != nil {
		return operations.NewGenericNotFound()
	}

	resp, err := http.Get(c.Photo)
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
	commenter, err := commenterGetByCommenterToken(*params.Body.CommenterToken)
	if err != nil {
		return operations.NewCommenterSelfOK().WithPayload(&operations.CommenterSelfOKBody{Message: err.Error()})
	}

	email, err := emailGet(commenter.Email)
	if err != nil {
		return operations.NewCommenterSelfOK().WithPayload(&operations.CommenterSelfOKBody{Message: err.Error()})
	}

	// Succeeded
	return operations.NewCommenterSelfOK().WithPayload(&operations.CommenterSelfOKBody{
		Commenter: commenter,
		Email:     email,
		Success:   true,
	})
}

func CommenterTokenNew(operations.CommenterTokenNewParams) middleware.Responder {
	commenterToken, err := commenterTokenNew()
	if err != nil {
		return operations.NewCommenterTokenNewOK().WithPayload(&operations.CommenterTokenNewOKBody{Message: err.Error()})
	}

	// Succeeded
	return operations.NewCommenterTokenNewOK().WithPayload(&operations.CommenterTokenNewOKBody{
		CommenterToken: commenterToken,
		Success:        true,
	})
}

func CommenterUpdate(params operations.CommenterUpdateParams) middleware.Responder {
	commenter, err := commenterGetByCommenterToken(*params.Body.CommenterToken)
	if err != nil {
		return operations.NewCommenterUpdateOK().WithPayload(&models.APIResponseBase{Message: err.Error()})
	}

	if commenter.Provider != "commento" {
		return operations.NewCommenterUpdateOK().WithPayload(&models.APIResponseBase{Message: util.ErrorCannotUpdateOauthProfile.Error()})
	}

	*params.Body.Email = commenter.Email

	if err = commenterUpdate(commenter.CommenterHex, *params.Body.Email, *params.Body.Name, params.Body.Link, params.Body.Photo, commenter.Provider); err != nil {
		return operations.NewCommenterUpdateOK().WithPayload(&models.APIResponseBase{Message: err.Error()})
	}

	// Succeeded
	return operations.NewCommenterUpdateOK().WithPayload(&models.APIResponseBase{Success: true})
}

func commenterGetByCommenterToken(commenterToken models.CommenterHexID) (*models.Commenter, error) {
	if commenterToken == "" {
		return nil, util.ErrorMissingField
	}

	row := svc.DB.QueryRow(
		fmt.Sprintf(
			"select %s from commentersessions "+
				"join commenters on commentersessions.commenterhex = commenters.commenterhex "+
				"where commentertoken = $1;",
			commentersRowColumns),
		commenterToken)

	var c models.Commenter
	if err := commentersRowScan(row, &c); err != nil {
		// TODO: is this the only error?
		return nil, util.ErrorNoSuchToken
	}

	if c.CommenterHex == "none" {
		return nil, util.ErrorNoSuchToken
	}

	return &c, nil
}

func commenterGetByEmail(provider string, email strfmt.Email) (*models.Commenter, error) {
	if provider == "" || email == "" {
		return nil, util.ErrorMissingField
	}
	row := svc.DB.QueryRow(
		fmt.Sprintf("select %s from commenters where email=$1 and provider=$2;", commentersRowColumns),
		email,
		provider,
	)

	var c models.Commenter
	if err := commentersRowScan(row, &c); err != nil {
		// TODO: is this the only error?
		return nil, util.ErrorNoSuchCommenter
	}

	return &c, nil
}

func commenterGetByHex(commenterHex models.CommenterHexID) (*models.Commenter, error) {
	if commenterHex == "" {
		return nil, util.ErrorMissingField
	}

	row := svc.DB.QueryRow(
		fmt.Sprintf("select %s from commenters where commenterHex = $1;", commentersRowColumns),
		commenterHex)
	var c models.Commenter
	if err := commentersRowScan(row, &c); err != nil {
		// TODO: is this the only error?
		return nil, util.ErrorNoSuchCommenter
	}

	return &c, nil
}

func commenterLogin(email strfmt.Email, password string) (models.CommenterHexID, error) {
	if email == "" || password == "" {
		return "", util.ErrorMissingField
	}

	row := svc.DB.QueryRow(
		"select commenterHex, passwordHash from commenters where email=$1 and provider='commento';",
		email)

	var commenterHex string
	var passwordHash string
	if err := row.Scan(&commenterHex, &passwordHash); err != nil {
		return "", util.ErrorInvalidEmailPassword
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		// TODO: is this the only possible error?
		return "", util.ErrorInvalidEmailPassword
	}

	commenterToken, err := util.RandomHex(32)
	if err != nil {
		logger.Errorf("cannot create commenterToken: %v", err)
		return "", util.ErrorInternal
	}

	_, err = svc.DB.Exec(
		"insert into commenterSessions(commenterToken, commenterHex, creationDate) values($1, $2, $3);",
		commenterToken,
		commenterHex,
		time.Now().UTC())
	if err != nil {
		logger.Errorf("cannot insert commenterToken token: %v\n", err)
		return "", util.ErrorInternal
	}

	return models.CommenterHexID(commenterToken), nil
}

func commenterNew(email strfmt.Email, name string, link string, photo string, provider string, password string) (models.CommenterHexID, error) {
	if email == "" || name == "" || link == "" || photo == "" || provider == "" {
		return "", util.ErrorMissingField
	}

	if provider == "commento" && password == "" {
		return "", util.ErrorMissingField
	}

	if link != "undefined" {
		if _, err := util.ParseAbsoluteURL(link); err != nil {
			return "", err
		}
	}

	if _, err := commenterGetByEmail(provider, email); err == nil {
		return "", util.ErrorEmailAlreadyExists
	}

	if err := EmailNew(email); err != nil {
		return "", util.ErrorInternal
	}

	commenterHex, err := util.RandomHex(32)
	if err != nil {
		return "", util.ErrorInternal
	}

	var passwordHash []byte
	if password != "" {
		passwordHash, err = bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			logger.Errorf("cannot generate hash from password: %v\n", err)
			return "", util.ErrorInternal
		}
	}

	statement := `insert into commenters(commenterHex, email, name, link, photo, provider, passwordHash, joinDate) values($1, $2, $3, $4, $5, $6, $7, $8);`
	_, err = svc.DB.Exec(statement, commenterHex, email, name, link, photo, provider, string(passwordHash), time.Now().UTC())
	if err != nil {
		logger.Errorf("cannot insert commenter: %v", err)
		return "", util.ErrorInternal
	}

	return models.CommenterHexID(commenterHex), nil
}

func commentersRowScan(s util.Scanner, c *models.Commenter) error {
	return s.Scan(
		&c.CommenterHex,
		&c.Email,
		&c.Name,
		&c.Link,
		&c.Photo,
		&c.Provider,
		&c.JoinDate,
	)
}

func commenterSessionUpdate(commenterToken models.HexID, commenterHex models.CommenterHexID) error {
	if commenterToken == "" || commenterHex == "" {
		return util.ErrorMissingField
	}

	if _, err := svc.DB.Exec("update commenterSessions set commenterHex = $2 where commenterToken = $1;", commenterToken, commenterHex); err != nil {
		logger.Errorf("error updating commenterHex: %v", err)
		return util.ErrorInternal
	}

	return nil
}

func commenterTokenNew() (models.CommenterHexID, error) {
	commenterToken, err := util.RandomHex(32)
	if err != nil {
		logger.Errorf("cannot create commenterToken: %v", err)
		return "", util.ErrorInternal
	}

	_, err = svc.DB.Exec(
		"insert into commenterSessions(commenterToken, creationDate) values($1, $2);",
		commenterToken,
		time.Now().UTC())
	if err != nil {
		logger.Errorf("cannot insert new commenterToken: %v", err)
		return "", util.ErrorInternal
	}

	return models.CommenterHexID(commenterToken), nil
}

func commenterUpdate(commenterHex models.CommenterHexID, email strfmt.Email, name string, link string, photo string, provider string) error {
	if email == "" || name == "" || provider == "" {
		return util.ErrorMissingField
	}

	// TODO ditch this "undefined" crap
	if link == "" {
		link = "undefined"
	}

	_, err := svc.DB.Exec(
		"update commenters set email=$3, name=$4, link=$5, photo=$6 where commenterHex=$1 and provider=$2;",
		commenterHex,
		provider,
		email,
		name,
		link,
		photo)
	if err != nil {
		logger.Errorf("cannot update commenter: %v", err)
		return util.ErrorInternal
	}

	return nil
}
