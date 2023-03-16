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
	// CreateCommenter creates and persists a new commenter. If no idp is provided, the local auth provider is assumed
	CreateCommenter(email, name, websiteURL, photoURL, idp, password string) (*data.UserCommenter, error)
	// CreateCommenterSession creates and persists a new commenter session record, returning session token
	CreateCommenterSession(id models.CommenterHexID) (models.CommenterHexID, error)
	// CreateResetToken creates and persists a new password reset token for the user of given kind ('entity') and hex ID
	CreateResetToken(userID models.HexID, entity models.Entity) (models.HexID, error)
	// DeleteOwnerByID removes an owner user by their hex ID
	DeleteOwnerByID(id models.HexID) error
	// DeleteResetTokens removes all password reset tokens for the given user
	DeleteResetTokens(userID models.HexID) error
	// FindCommenterByID finds and returns a commenter user by their hex ID
	FindCommenterByID(id models.CommenterHexID) (*data.UserCommenter, error)
	// FindCommenterByIdPEmail finds and returns a commenter user by their email and identity provider. If no idp is
	// provided, the local auth provider (Comentario) is assumed
	FindCommenterByIdPEmail(idp, email string, readPwdHash bool) (*data.UserCommenter, error)
	// FindCommenterByToken finds and returns a commenter user by their token
	FindCommenterByToken(token models.CommenterHexID) (*data.UserCommenter, error)
	// FindOwnerByEmail finds and returns an owner user by their email
	FindOwnerByEmail(email string, readPwdHash bool) (*data.UserOwner, error)
	// FindOwnerByID finds and returns an owner user by their hex ID
	FindOwnerByID(id models.HexID) (*data.UserOwner, error)
	// FindOwnerByToken finds and returns an owner user by their token
	FindOwnerByToken(token models.HexID) (*data.UserOwner, error)
	// ResetUserPasswordByToken finds and resets a user's password for the given reset token, returning the
	// corresponding entity
	ResetUserPasswordByToken(token models.HexID, password string) (models.Entity, error)
	// UpdateCommenter updates the given commenter's data in the database. If no idp is provided, the local auth
	// provider is assumed
	UpdateCommenter(commenterHex models.CommenterHexID, email, name, websiteURL, photoURL, idp string) error
	// UpdateCommenterSession links a commenter token to the given commenter, by updating the session record
	UpdateCommenterSession(token, id models.CommenterHexID) error
}

//----------------------------------------------------------------------------------------------------------------------

// userService is a blueprint UserService implementation
type userService struct{}

func (svc *userService) CreateCommenter(email, name, websiteURL, photoURL, idp, password string) (*data.UserCommenter, error) {
	logger.Debugf("userService.CreateCommenter(%s, %s, %s, %s, %s, %s)", email, name, websiteURL, photoURL, idp, password)

	// Verify no such email is registered yet
	idp = idpFix(idp)
	if _, err := svc.FindCommenterByIdPEmail(idp, email, false); err == nil {
		return nil, ErrDuplicateEmail
	} else if err != ErrNotFound {
		return nil, err
	}

	// Register a new email
	if _, err := TheEmailService.Create(email); err != nil {
		return nil, err
	}

	// Create an initial commenter instance
	uc := data.UserCommenter{
		User: data.User{
			Email:   email,
			Created: time.Now().UTC(),
			Name:    name,
		},
		WebsiteURL: websiteURL,
		PhotoURL:   photoURL,
		Provider:   idp,
	}

	// Generate a random hex ID
	if id, err := data.RandomHexID(); err != nil {
		return nil, err
	} else {
		uc.HexID = id
	}

	// Hash the user's password, if any
	if password != "" {
		if h, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost); err != nil {
			return nil, err
		} else {
			uc.PasswordHash = string(h)
		}
	}

	// Insert a commenter record
	err := db.Exec(
		"insert into commenters(commenterHex, email, name, link, photo, provider, passwordHash, joinDate) values($1, $2, $3, $4, $5, $6, $7, $8);",
		uc.HexID,
		uc.Email,
		uc.Name,
		uc.WebsiteURL,
		uc.PhotoURL,
		uc.Provider,
		uc.PasswordHash,
		time.Now().UTC())
	if err != nil {
		logger.Errorf("userService.CreateCommenter: Exec() failed: %v", err)
		return nil, translateErrors(err)
	}

	// Succeeded
	return &uc, nil
}

func (svc *userService) CreateCommenterSession(id models.CommenterHexID) (models.CommenterHexID, error) {
	logger.Debugf("userService.CreateCommenterSession(%s)", id)

	// Generate a new random token
	token, err := data.RandomHexID()
	if err != nil {
		logger.Errorf("userService.CreateCommenterSession: RandomHexID() failed: %v", err)
		return "", err
	}

	// Insert a new record
	err = db.Exec(
		"insert into commentersessions(commentertoken, commenterhex, creationdate) values($1, $2, $3);",
		token, id, time.Now().UTC())
	if err != nil {
		logger.Errorf("userService.CreateCommenterSession: Exec() failed: %v", err)
		return "", translateErrors(err)
	}

	// Succeeded
	return models.CommenterHexID(token), nil
}

func (svc *userService) CreateResetToken(userID models.HexID, entity models.Entity) (models.HexID, error) {
	logger.Debugf("userService.CreateResetToken(%s, %s)", userID, entity)

	// Generate a random reset token
	token, err := data.RandomHexID()
	if err != nil {
		logger.Errorf("userService.CreateResetToken: util.RandomHexID() failed: %v", err)
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

func (svc *userService) DeleteOwnerByID(id models.HexID) error {
	logger.Debugf("userService.DeleteOwnerByID(%s)", id)

	// Remove all user's reset tokens
	if err := svc.DeleteResetTokens(id); err != nil {
		return err
	}

	// Remove all user's sessions
	if err := db.Exec("delete from ownersessions where ownerhex=$1;", id); err != nil {
		logger.Errorf("userService.DeleteOwnerByID: Exec() failed for ownersessions: %v", err)
		return translateErrors(err)
	}

	// Delete the owner user
	if err := db.Exec("delete from owners where ownerhex=$1;", id); err != nil {
		logger.Errorf("userService.DeleteOwnerByID: Exec() failed for owners: %v", err)
		return translateErrors(err)
	}

	// Succeeded
	return nil
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
		"select commenterhex, email, name, link, photo, provider, joindate, passwordhash from commenters where commenterhex=$1;",
		id)

	// Fetch the commenter user
	if u, err := svc.fetchCommenter(row, false); err != nil {
		return nil, translateErrors(err)
	} else {
		return u, nil
	}
}

func (svc *userService) FindCommenterByIdPEmail(idp, email string, readPwdHash bool) (*data.UserCommenter, error) {
	logger.Debugf("userService.FindCommenterByIdPEmail(%s, %s)", idp, email)

	// Query the database
	row := db.QueryRow(
		"select commenterhex, email, name, link, photo, provider, joindate, passwordhash "+
			"from commenters "+
			"where provider=$1 and email=$2;",
		idpFix(idp),
		email)

	// Fetch the commenter user
	if u, err := svc.fetchCommenter(row, readPwdHash); err != nil {
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
		"select c.commenterHex, c.email, c.name, c.link, c.photo, c.provider, c.joinDate, c.passwordhash "+
			"from commentersessions s "+
			"join commenters c on s.commenterhex = c.commenterhex "+
			"where s.commentertoken=$1;",
		token)

	// Fetch the commenter user
	if u, err := svc.fetchCommenter(row, false); err != nil {
		return nil, translateErrors(err)
	} else {
		return u, nil
	}
}

func (svc *userService) FindOwnerByEmail(email string, readPwdHash bool) (*data.UserOwner, error) {
	logger.Debugf("userService.FindOwnerByEmail(%s)", email)

	// Query the database
	row := db.QueryRow("select ownerHex, email, name, confirmedEmail, joinDate, passwordhash from owners where email=$1;", email)

	// Fetch the owner user
	if u, err := svc.fetchOwner(row, readPwdHash); err != nil {
		return nil, translateErrors(err)
	} else {
		return u, nil
	}
}

func (svc *userService) FindOwnerByID(id models.HexID) (*data.UserOwner, error) {
	logger.Debugf("userService.FindOwnerByID(%s)", id)

	// Query the database
	row := db.QueryRow("select ownerHex, email, name, confirmedEmail, joinDate, passwordhash from owners where ownerhex=$1;", id)

	// Fetch the owner user
	if u, err := svc.fetchOwner(row, false); err != nil {
		return nil, translateErrors(err)
	} else {
		return u, nil
	}
}

func (svc *userService) FindOwnerByToken(token models.HexID) (*data.UserOwner, error) {
	logger.Debugf("userService.FindOwnerByToken(%s)", token)

	// Query the database
	row := db.QueryRow(
		"select ownerHex, email, name, confirmedEmail, joinDate, passwordhash "+
			"from owners "+
			"where ownerHex in (select ownerHex from ownersessions where ownertoken=$1);",
		token)

	// Fetch the owner user
	if u, err := svc.fetchOwner(row, false); err != nil {
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

func (svc *userService) UpdateCommenter(commenterHex models.CommenterHexID, email, name, websiteURL, photoURL, idp string) error {
	logger.Debugf("userService.UpdateCommenter(%s, %s, %s, %s, %s, %s)", commenterHex, email, name, websiteURL, photoURL, idp)

	// TODO ditch this "undefined" crap
	if websiteURL == "" {
		websiteURL = "undefined"
	}

	// Update the database record
	err := db.Exec(
		"update commenters set email=$1, name=$2, link=$3, photo=$4 where commenterhex=$5 and provider=$6;",
		email, name, websiteURL, photoURL, commenterHex, idpFix(idp))
	if err != nil {
		logger.Errorf("userService.UpdateCommenter: Exec() failed: %v", err)
		return translateErrors(err)
	}

	// Succeeded
	return nil
}

func (svc *userService) UpdateCommenterSession(token, id models.CommenterHexID) error {
	logger.Debugf("userService.UpdateCommenterSession(%s, %s)", token, id)

	// Update the record
	if err := db.Exec("update commentersessions set commenterhex=$1 where commentertoken=$2;", id, token); err != nil {
		logger.Errorf("userService.UpdateCommenterSession: Exec() failed: %v", err)
		return translateErrors(err)
	}

	// Succeeded
	return nil
}

// fetchCommenter returns a new commenter user from the provided database row
func (svc *userService) fetchCommenter(s util.Scanner, readPwdHash bool) (*data.UserCommenter, error) {
	u := data.UserCommenter{}
	var pwdHash string
	if err := s.Scan(&u.HexID, &u.Email, &u.Name, &u.WebsiteURL, &u.PhotoURL, &u.Provider, &u.Created, &pwdHash); err != nil {
		// Log "not found" errors only in debug
		if err != sql.ErrNoRows || logger.IsEnabledFor(logging.DEBUG) {
			logger.Errorf("userService.fetchCommenter: Scan() failed: %v", err)
		}
		return nil, err
	}
	if readPwdHash {
		u.PasswordHash = pwdHash
	}
	return &u, nil
}

// fetchOwner returns a new owner user instance from the provided database row
func (svc *userService) fetchOwner(s util.Scanner, readPwdHash bool) (*data.UserOwner, error) {
	u := data.UserOwner{}
	var pwdHash string
	if err := s.Scan(&u.HexID, &u.Email, &u.Name, &u.EmailConfirmed, &u.Created, &pwdHash); err != nil {
		// Log "not found" errors only in debug
		if err != sql.ErrNoRows || logger.IsEnabledFor(logging.DEBUG) {
			logger.Errorf("userService.fetchOwner: Scan() failed: %v", err)
		}
		return nil, err
	}
	if readPwdHash {
		u.PasswordHash = pwdHash
	}
	return &u, nil
}

// idpFix handles default value for IdP
func idpFix(idp string) string {
	// IdP defaults to local
	if idp == "" {
		return "commento"
	}
	return idp
}
