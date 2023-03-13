package svc

import (
	"gitlab.com/comentario/comentario/internal/api/models"
	"gitlab.com/comentario/comentario/internal/data"
	"gitlab.com/comentario/comentario/internal/util"
)

// TheUserService is a global UserService implementation
var TheUserService UserService = &userService{}

// UserService is a service interface for dealing with users
type UserService interface {
	// FindOwnerByEmail finds and returns an owner user by their email
	FindOwnerByEmail(email string) (*data.User, error)
	// FindOwnerByToken finds and returns an owner user by their token
	FindOwnerByToken(token models.HexID) (*data.User, error)
}

//----------------------------------------------------------------------------------------------------------------------

// userService is a blueprint UserService implementation
type userService struct{}

func (svc *userService) FindOwnerByEmail(email string) (*data.User, error) {
	// Validate the passed email
	if err := validateEmail(email); err != nil {
		return nil, err
	}

	// Query the database
	row := db.QueryRow("select ownerHex, email, name, confirmedEmail, joinDate from owners where email=$1;", email)

	// Fetch the user
	if u, err := svc.fetchOwnerUser(row); err != nil {
		return nil, translateErrors(err)
	} else {
		return u, nil
	}
}

func (svc *userService) FindOwnerByToken(token models.HexID) (*data.User, error) {
	// Validate the passed token
	if err := validateHexID(token); err != nil {
		return nil, err
	}

	// Query the database
	row := db.QueryRow(
		"select ownerHex, email, name, confirmedEmail, joinDate "+
			"from owners "+
			"where ownerHex in (select ownerHex from ownersessions where ownerToken=$1);",
		token)

	// Fetch the user
	if u, err := svc.fetchOwnerUser(row); err != nil {
		return nil, translateErrors(err)
	} else {
		return u, nil
	}
}

// fetchOwnerUser returns a user of kind Owner from the provided database row
func (svc *userService) fetchOwnerUser(s util.Scanner) (*data.User, error) {
	u := data.User{Kind: data.UserKindOwner}
	if err := s.Scan(&u.HexID, &u.Email, &u.Name, &u.EmailConfirmed, &u.Created); err != nil {
		logger.Errorf("userService.fetchOwnerUser: Scan() failed: %v", err)
		return nil, err
	}
	return &u, nil
}
