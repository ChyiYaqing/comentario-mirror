package svc

import (
	"gitlab.com/comentario/comentario/internal/api/models"
	"gitlab.com/comentario/comentario/internal/util"
)

// TheEmailService is a global EmailService implementation
var TheEmailService EmailService = &emailService{}

// EmailService is a service interface for dealing with email objects
type EmailService interface {
	// FindByEmail finds and returns an Email instance for the given email address
	FindByEmail(email string) (*models.Email, error)
	// FindByUnsubscribeToken finds and returns an Email instance for the given unsubscribe token
	FindByUnsubscribeToken(token models.HexID) (*models.Email, error)
	// UpdateByEmailToken updates an Email instance for the given email address and unsubscribe token
	UpdateByEmailToken(email string, token models.HexID, sendReply, sendModerator bool) error
}

//----------------------------------------------------------------------------------------------------------------------

// emailService is a blueprint EmailService implementation
type emailService struct{}

func (svc *emailService) FindByEmail(email string) (*models.Email, error) {
	logger.Debugf("emailService.FindByEmail(%s)", email)

	// Query the database row
	row := db.QueryRow(
		"select email, unsubscribesecrethex, lastemailnotificationdate, sendreplynotifications, sendmoderatornotifications "+
			"from emails "+
			"where email=$1;",
		email)

	// Fetch the email
	if e, err := svc.fetchEmail(row); err != nil {
		return nil, translateErrors(err)
	} else {
		return e, nil
	}
}

func (svc *emailService) FindByUnsubscribeToken(token models.HexID) (*models.Email, error) {
	logger.Debugf("emailService.FindByUnsubscribeToken(%s)", token)

	// Query the database row
	row := db.QueryRow(
		"select email, unsubscribesecrethex, lastemailnotificationdate, sendreplynotifications, sendmoderatornotifications "+
			"from emails "+
			"where unsubscribesecrethex=$1;",
		token)

	// Fetch the email
	if e, err := svc.fetchEmail(row); err != nil {
		return nil, translateErrors(err)
	} else {
		return e, nil
	}
}

func (svc *emailService) UpdateByEmailToken(email string, token models.HexID, sendReply, sendModerator bool) error {
	logger.Debugf("emailService.UpdateByEmailToken(%s)", token)

	// Update the database row
	err := db.Exec(
		"update emails set sendreplynotifications=$1, sendmoderatornotifications=$2 where email=$3 and unsubscribesecrethex=$4;",
		sendReply,
		sendModerator,
		email,
		token)
	if err != nil {
		logger.Errorf("emailService.UpdateByEmailToken: Exec() failed: %v", err)
		return translateErrors(err)
	}

	// Succeeded
	return nil
}

// fetchEmail returns a new Email instance from the provided database row
func (svc *emailService) fetchEmail(s util.Scanner) (*models.Email, error) {
	e := models.Email{}
	if err := s.Scan(&e.Email, &e.UnsubscribeSecretHex, &e.LastEmailNotificationDate, &e.SendReplyNotifications, &e.SendModeratorNotifications); err != nil {
		logger.Errorf("emailService.fetchEmail: Scan() failed: %v", err)
		return nil, err
	}
	return &e, nil
}
