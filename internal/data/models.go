package data

import (
	"github.com/go-openapi/strfmt"
	"gitlab.com/comentario/comentario/internal/api/models"
	"time"
)

const RootParentHexID = models.ParentHexID("root") // The "root" parent hex

// AnonymousCommenter is a fake, anonymous, commenter instance, which doesn't exist in the database, but is nonetheless
// referenced by comments ¯\_(ツ)_/¯
var AnonymousCommenter = UserCommenter{
	User: User{
		HexID: "0000000000000000000000000000000000000000000000000000000000000000",
		Name:  "Anonymous",
	},
}

// Principal represents user's identity for the API
type Principal interface {
	// GetHexID returns the underlying user's hex ID
	GetHexID() models.HexID
	// GetUser returns the underlying User instance
	GetUser() *User
	// IsAnonymous returns whether the underlying user is anonymous
	IsAnonymous() bool
}

// User is a base user type
type User struct {
	HexID        models.HexID // User hex ID
	Email        string       // User's email
	Created      time.Time    // Timestamp when user was created, in UTC
	Name         string       // User's full name
	PasswordHash string       // User's hashed password
}

func (u *User) GetHexID() models.HexID {
	return u.HexID
}

func (u *User) GetUser() *User {
	return u
}

func (u *User) IsAnonymous() bool {
	return u.HexID == AnonymousCommenter.HexID
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

// ToCommenter converts this user into models.Commenter model
func (u *UserCommenter) ToCommenter() *models.Commenter {
	return &models.Commenter{
		CommenterHex: u.HexID,
		Email:        strfmt.Email(u.Email),
		IsModerator:  u.IsModerator,
		JoinDate:     strfmt.DateTime(u.Created),
		Link:         u.WebsiteURL,
		Name:         u.Name,
		Photo:        u.PhotoURL,
		Provider:     u.Provider,
	}
}
