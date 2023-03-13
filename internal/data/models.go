package data

import (
	"github.com/go-openapi/strfmt"
	"gitlab.com/comentario/comentario/internal/api/models"
	"time"
)

// UserKind denotes what kind a user represents
type UserKind int

const (
	UserKindAdmin     UserKind = iota // Admin
	UserKindOwner                     // Owner
	UserKindCommenter                 // Commenter
)

type User struct {
	Kind           UserKind     // User kind
	HexID          models.HexID // User hex ID
	Email          string       // User's email
	EmailConfirmed bool         // Whether the user's email is confirmed
	Created        time.Time    // Timestamp when user was created, in UTC
	Name           string       // User's full name
}

// ToOwner converts this user into models.Owner model
func (u *User) ToOwner() *models.Owner {
	// Verify the user is indeed an Owner
	if u.Kind != UserKindOwner {
		panic("user is not of kind Owner")
	}

	// Convert the model
	return &models.Owner{
		ConfirmedEmail: u.EmailConfirmed,
		Email:          strfmt.Email(u.Email),
		JoinDate:       strfmt.DateTime(u.Created),
		Name:           u.Name,
		OwnerHex:       u.HexID,
	}
}
