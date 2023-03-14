package data

import (
	"github.com/go-openapi/strfmt"
	"gitlab.com/comentario/comentario/internal/api/models"
	"time"
)

const RootParentHexID = models.ParentHexID("root")                 // The "root" parent hex
const AnonymousCommenterHexID = models.CommenterHexID("anonymous") // The "anonymous" commenter hex ID or token

// User is a base user type
type User struct {
	HexID   models.HexID // User hex ID
	Email   string       // User's email
	Created time.Time    // Timestamp when user was created, in UTC
	Name    string       // User's full name
}

// ---------------------------------------------------------------------------------------------------------------------

// UserOwner represents a user that is a domain owner
type UserOwner struct {
	User
	EmailConfirmed bool // Whether the user's email is confirmed
}

// ToOwner converts this user into models.Owner model
func (u *UserOwner) ToOwner() *models.Owner {
	return &models.Owner{
		ConfirmedEmail: u.EmailConfirmed,
		Email:          strfmt.Email(u.Email),
		JoinDate:       strfmt.DateTime(u.Created),
		Name:           u.Name,
		OwnerHex:       u.HexID,
	}
}

// ---------------------------------------------------------------------------------------------------------------------

// UserCommenter represents a commenter user
type UserCommenter struct {
	User
	IsModerator bool   // Whether the user is a moderator
	WebsiteURL  string // User's website link
	PhotoURL    string // URL of the user's avatar image
	Provider    string // User's federated provider ID
}

// CommenterHexID returns the ID of the user converted into a CommenterHexID
func (u *UserCommenter) CommenterHexID() models.CommenterHexID {
	return models.CommenterHexID(u.HexID)
}

// ToCommenter converts this user into models.Commenter model
func (u *UserCommenter) ToCommenter() *models.Commenter {
	return &models.Commenter{
		CommenterHex: u.CommenterHexID(),
		Email:        strfmt.Email(u.Email),
		IsModerator:  u.IsModerator,
		JoinDate:     strfmt.DateTime(u.Created),
		Link:         u.WebsiteURL,
		Name:         u.Name,
		Photo:        u.PhotoURL,
		Provider:     u.Provider,
	}
}
