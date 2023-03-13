package svc

import (
	"gitlab.com/comentario/comentario/internal/api/models"
	"time"
)

// TheVoteService is a global VoteService implementation
var TheVoteService VoteService = &voteService{}

// VoteService is a service interface for dealing with comment votes
type VoteService interface {
	// SetVote inserts or updates a vote for the given comment and commenter
	SetVote(commentHex models.HexID, commenterHex models.CommenterHexID, direction int) error
}

//----------------------------------------------------------------------------------------------------------------------

// voteService is a blueprint VoteService implementation
type voteService struct{}

func (svc *voteService) SetVote(commentHex models.HexID, commenterHex models.CommenterHexID, direction int) error {
	// Validate the IDs
	if err := validateHexID(commentHex); err != nil {
		return err
	}

	// Validate the direction
	if direction < -1 || direction > 1 {
		return ErrInvalidInput
	}

	// Upsert a row
	_, err := db.Exec(
		"insert into votes(commentHex, commenterHex, direction, voteDate) values($1, $2, $3, $4) "+
			"on conflict (commentHex, commenterHex) do update set direction = $3;",
		commentHex,
		commenterHex,
		direction,
		time.Now().UTC())
	if err != nil {
		logger.Errorf("voteService.SetVote(): Exec failed: %v", err)
		return translateErrors(err)
	}

	// Succeeded
	return nil
}
