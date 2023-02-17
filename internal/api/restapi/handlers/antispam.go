package handlers

import (
	"github.com/adtac/go-akismet/akismet"
	"gitlab.com/comentario/comentario/internal/config"
)

func checkForSpam(domain string, userIp string, userAgent string, name string, email string, url string, markdown string) bool {
	// Ignore if Akismet isn't configured
	if config.SecretsConfig.Akismet.Key == "" {
		return false
	}

	res, err := akismet.Check(
		&akismet.Comment{
			Blog:               domain,
			UserIP:             userIp,
			UserAgent:          userAgent,
			CommentType:        "comment",
			CommentAuthor:      name,
			CommentAuthorEmail: email,
			CommentAuthorURL:   url,
			CommentContent:     markdown,
		},
		config.SecretsConfig.Akismet.Key)

	if err != nil {
		logger.Errorf("error: cannot validate comment using Akismet: %v", err)
		return true
	}

	return res
}
