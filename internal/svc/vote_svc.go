package svc

import (
	"gitlab.com/comentario/comentario/internal/api/models"
	"time"
)

// TheVoteService is a global VoteService implementation
var TheVoteService VoteService = &voteService{}

// VoteService is a service interface for dealing with comment votes
type VoteService interface {
	// DeleteByDomain deletes all votes for the specified domain
	DeleteByDomain(domain string) error
	// SetVote inserts or updates a vote for the given comment and commenter
	SetVote(commentHex models.HexID, commenterHex models.CommenterHexID, direction int) error
}

//----------------------------------------------------------------------------------------------------------------------

// voteService is a blueprint VoteService implementation
type voteService struct{}

func (svc *voteService) DeleteByDomain(domain string) error {
	logger.Debugf("voteService.DeleteByDomain(%s)", domain)

	// Delete the records in the database
	if err := db.Exec("delete from votes v using comments c where c.commentHex=v.commentHex and c.domain=$1;", domain); err != nil {
		logger.Errorf("voteService.DeleteByDomain: Exec() failed: %v", err)
		return translateErrors(err)
	}

	// Succeeded
	return nil
}

func (svc *voteService) SetVote(commentHex models.HexID, commenterHex models.CommenterHexID, direction int) error {
	logger.Debugf("voteService.SetVote(%s, %s, %d)", commentHex, commenterHex, direction)

	// Upsert a row
	err := db.Exec(
		"insert into votes(commentHex, commenterHex, direction, voteDate) values($1, $2, $3, $4) "+
			"on conflict (commentHex, commenterHex) do update set direction = $3;",
		commentHex,
		commenterHex,
		direction,
		time.Now().UTC())
	if err != nil {
		logger.Errorf("voteService.SetVote: Exec() failed: %v", err)
		return translateErrors(err)
	}

	// Succeeded
	return nil
}
