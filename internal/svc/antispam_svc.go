package svc

import (
	"github.com/adtac/go-akismet/akismet"
	"gitlab.com/comentario/comentario/internal/config"
)

// TheAntispamService is a global AntispamService implementation
var TheAntispamService AntispamService = &antispamService{}

// AntispamService is a service interface for spam checks
type AntispamService interface {
	// CheckForSpam verifies the provided message details for spam and returns whether it's a spam message
	CheckForSpam(domain, userIP, userAgent, name, email, url, markdown string) bool
}

//----------------------------------------------------------------------------------------------------------------------

// antispamService is a blueprint AntispamService implementation
type antispamService struct{}

func (svc *antispamService) CheckForSpam(domain, userIP, userAgent, name, email, url, markdown string) bool {
	logger.Debugf("antispamService.CheckForSpam(%s, %s, %s, %s, %s, ...)", domain, userIP, userAgent, name, email, url)

	// Ignore if Akismet isn't configured (consider it not spam)
	if config.SecretsConfig.Akismet.Key == "" {
		return false
	}

	// Run the message with Akismet API
	res, err := akismet.Check(
		&akismet.Comment{
			Blog:               domain,
			UserIP:             userIP,
			UserAgent:          userAgent,
			CommentType:        "comment",
			CommentAuthor:      name,
			CommentAuthorEmail: email,
			CommentAuthorURL:   url,
			CommentContent:     markdown,
		},
		config.SecretsConfig.Akismet.Key)
	if err != nil {
		logger.Errorf("antispamService.CheckForSpam: akismet.Check() failed: %v", err)
		return true
	}
	return res
}
