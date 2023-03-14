package svc

import (
	"database/sql"
	"github.com/op/go-logging"
	"gitlab.com/comentario/comentario/internal/api/models"
	"gitlab.com/comentario/comentario/internal/data"
	"gitlab.com/comentario/comentario/internal/util"
	"golang.org/x/crypto/bcrypt"
	"time"
)

// TheUserService is a global UserService implementation
var TheUserService UserService = &userService{}

// UserService is a service interface for dealing with users
type UserService interface {
	// CreateResetToken creates and persists a new password reset token for the user of given kind ('entity') and hex ID
	CreateResetToken(userID models.HexID, entity models.Entity) (string, error)
	// DeleteResetTokens removes all password reset tokens for the given user
	DeleteResetTokens(userID models.HexID) error
	// FindCommenterByID finds and returns a commenter user by their hex ID
	FindCommenterByID(id models.CommenterHexID) (*data.UserCommenter, error)
	// FindCommenterByIdPEmail finds and returns a commenter user by their email and identity provider. If no idp is
	// provided, the local auth provider (Comentario) is assumed
	FindCommenterByIdPEmail(idp, email string) (*data.UserCommenter, error)
	// FindCommenterByToken finds and returns a commenter user by their token
	FindCommenterByToken(token models.CommenterHexID) (*data.UserCommenter, error)
	// FindOwnerByEmail finds and returns an owner user by their email
	FindOwnerByEmail(email string) (*data.UserOwner, error)
	// FindOwnerByID finds and returns an owner user by their hex ID
	FindOwnerByID(id models.HexID) (*data.UserOwner, error)
	// FindOwnerByToken finds and returns an owner user by their token
	FindOwnerByToken(token models.HexID) (*data.UserOwner, error)
	// ResetUserPasswordByToken finds and resets a user's password for the given reset token, returning the
	// corresponding entity
	ResetUserPasswordByToken(token models.HexID, password string) (models.Entity, error)
}

//----------------------------------------------------------------------------------------------------------------------

// userService is a blueprint UserService implementation
type userService struct{}

func (svc *userService) CreateResetToken(userID models.HexID, entity models.Entity) (string, error) {
	logger.Debugf("userService.CreateResetToken(%s, %s)", userID, entity)

	// Generate a random reset token
	token, err := util.RandomHex(32)
	if err != nil {
		logger.Errorf("userService.CreateResetToken: util.RandomHex() failed: %v", err)
		return "", err
	}

	// Persist the token
	err = db.Exec(
		"insert into resetHexes(resetHex, hex, entity, sendDate) values($1, $2, $3, $4);",
		token,
		userID,
		entity,
		time.Now().UTC())
	if err != nil {
		logger.Errorf("userService.CreateResetToken: Exec() failed: %v", err)
		return "", translateErrors(err)
	}

	// Succeeded
	return token, nil
}

func (svc *userService) DeleteResetTokens(userID models.HexID) error {
	logger.Debugf("userService.DeleteResetTokens(%s)", userID)

	// Delete all tokens by user
	if err := db.Exec("delete from resethexes where hex=$1;", userID); err != nil {
		logger.Errorf("userService.DeleteResetTokens: Exec() failed: %v", err)
		return translateErrors(err)
	}

	// Succeeded
	return nil
}

func (svc *userService) FindCommenterByID(id models.CommenterHexID) (*data.UserCommenter, error) {
	logger.Debugf("userService.FindCommenterByID(%s)", id)

	// Make sure we don't try to find an "anonymous" commenter
	if id == data.AnonymousCommenterHexID {
		return nil, ErrNotFound
	}

	// Query the database
	row := db.QueryRow(
		"select commenterHex, email, name, link, photo, provider, joinDate from commenters where commenterhex=$1;",
		id)

	// Fetch the commenter user
	if u, err := svc.fetchCommenter(row); err != nil {
		return nil, translateErrors(err)
	} else {
		return u, nil
	}
}

func (svc *userService) FindCommenterByIdPEmail(idp, email string) (*data.UserCommenter, error) {
	logger.Debugf("userService.FindCommenterByIdPEmail(%s, %s)", idp, email)

	// IdP defaults to local
	if idp == "" {
		idp = "commento"
	}

	// Query the database
	row := db.QueryRow(
		"select commenterHex, email, name, link, photo, provider, joinDate from commenters where provider=$1 and email=$2;",
		idp,
		email)

	// Fetch the commenter user
	if u, err := svc.fetchCommenter(row); err != nil {
		return nil, translateErrors(err)
	} else {
		return u, nil
	}
}

func (svc *userService) FindCommenterByToken(token models.CommenterHexID) (*data.UserCommenter, error) {
	logger.Debugf("userService.FindCommenterByToken(%s)", token)

	// Make sure we don't try to find an "anonymous" commenter
	if token == data.AnonymousCommenterHexID {
		return nil, ErrNotFound
	}

	// Query the database
	row := db.QueryRow(
		"select c.commenterHex, c.email, c.name, c.link, c.photo, c.provider, c.joinDate "+
			"from commentersessions s "+
			"join commenters c on s.commenterhex = c.commenterhex "+
			"where s.commentertoken=$1;",
		token)

	// Fetch the commenter user
	if u, err := svc.fetchCommenter(row); err != nil {
		return nil, translateErrors(err)
	} else {
		return u, nil
	}
}

func (svc *userService) FindOwnerByEmail(email string) (*data.UserOwner, error) {
	logger.Debugf("userService.FindOwnerByEmail(%s)", email)

	// Query the database
	row := db.QueryRow("select ownerHex, email, name, confirmedEmail, joinDate from owners where email=$1;", email)

	// Fetch the owner user
	if u, err := svc.fetchOwner(row); err != nil {
		return nil, translateErrors(err)
	} else {
		return u, nil
	}
}

func (svc *userService) FindOwnerByID(id models.HexID) (*data.UserOwner, error) {
	logger.Debugf("userService.FindOwnerByID(%s)", id)

	// Query the database
	row := db.QueryRow("select ownerHex, email, name, confirmedEmail, joinDate from owners where ownerhex=$1;", id)

	// Fetch the owner user
	if u, err := svc.fetchOwner(row); err != nil {
		return nil, translateErrors(err)
	} else {
		return u, nil
	}
}

func (svc *userService) FindOwnerByToken(token models.HexID) (*data.UserOwner, error) {
	logger.Debugf("userService.FindOwnerByToken(%s)", token)

	// Query the database
	row := db.QueryRow(
		"select ownerHex, email, name, confirmedEmail, joinDate "+
			"from owners "+
			"where ownerHex in (select ownerHex from ownersessions where ownertoken=$1);",
		token)

	// Fetch the owner user
	if u, err := svc.fetchOwner(row); err != nil {
		return nil, translateErrors(err)
	} else {
		return u, nil
	}
}

func (svc *userService) ResetUserPasswordByToken(token models.HexID, password string) (models.Entity, error) {
	logger.Debugf("userService.ResetUserPasswordByToken(%s, %s)", token, password)

	// Find and fetch the token record
	var userID models.HexID
	var entity models.Entity
	row := db.QueryRow("select hex, entity from resethexes where resethex=$1;", token)
	if err := row.Scan(&userID, &entity); err != nil {
		// Do not log "not found" errors
		if err != sql.ErrNoRows {
			logger.Errorf("userService.ResetUserPasswordByToken: Scan() failed: %v", err)
		}
		return "", translateErrors(err)
	}

	// Hash the new password
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		logger.Errorf("cannot generate hash from password: %v", err)
		return "", err
	}

	// Fetch the user and update their password
	switch entity {
	// Owner
	case models.EntityOwner:
		if _, err := svc.FindOwnerByID(userID); err != nil {
			return "", err
		} else if err := db.Exec("update owners set passwordhash=$1 where ownerhex=$2;", string(hash), userID); err != nil {
			logger.Errorf("userService.ResetUserPasswordByToken: Exec() failed for owner: %v", err)
			return "", translateErrors(err)
		}

	// Commenter
	case models.EntityCommenter:
		if _, err := svc.FindCommenterByID(models.CommenterHexID(userID)); err != nil {
			return "", err
		} else if err := db.Exec("update commenters set passwordhash=$1 where commenterhex=$2;", string(hash), userID); err != nil {
			logger.Errorf("userService.ResetUserPasswordByToken: Exec() failed for commenter: %v", err)
			return "", translateErrors(err)
		}

	// Unknown entity
	default:
		return "", ErrUnknownEntity
	}

	// Remove all the user's reset tokens, ignoring any error
	_ = svc.DeleteResetTokens(userID)

	// Succeeded
	return entity, nil
}

// fetchCommenter returns a new commenter user from the provided database row
func (svc *userService) fetchCommenter(s util.Scanner) (*data.UserCommenter, error) {
	u := data.UserCommenter{}
	if err := s.Scan(&u.HexID, &u.Email, &u.Name, &u.WebsiteURL, &u.PhotoURL, &u.Provider, &u.Created); err != nil {
		// Log "not found" errors only in debug
		if err != sql.ErrNoRows || logger.IsEnabledFor(logging.DEBUG) {
			logger.Errorf("userService.fetchCommenter: Scan() failed: %v", err)
		}
		return nil, err
	}
	return &u, nil
}

// fetchOwner returns a new owner user instance from the provided database row
func (svc *userService) fetchOwner(s util.Scanner) (*data.UserOwner, error) {
	u := data.UserOwner{}
	if err := s.Scan(&u.HexID, &u.Email, &u.Name, &u.EmailConfirmed, &u.Created); err != nil {
		// Log "not found" errors only in debug
		if err != sql.ErrNoRows || logger.IsEnabledFor(logging.DEBUG) {
			logger.Errorf("userService.fetchOwner: Scan() failed: %v", err)
		}
		return nil, err
	}
	return &u, nil
}
