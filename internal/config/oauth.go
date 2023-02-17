package config

import (
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/gitlab"
	"golang.org/x/oauth2/google"
)

var (
	OAuthGithubConfig *oauth2.Config
	OAuthGitlabConfig *oauth2.Config
	OAuthGoogleConfig *oauth2.Config
)

func oauthConfigure() {
	githubOauthConfigure()
	gitlabOauthConfigure()
	googleOauthConfigure()
}

func githubOauthConfigure() {
	if !SecretsConfig.IdP.GitHub.Usable() {
		logger.Debug("GitHub auth isn't configured or enabled")
		OAuthGithubConfig = nil
	}

	OAuthGithubConfig = &oauth2.Config{
		RedirectURL:  URLForAPI("oauth/github/callback", nil),
		ClientID:     SecretsConfig.IdP.GitHub.Key,
		ClientSecret: SecretsConfig.IdP.GitHub.Secret,
		Scopes:       []string{"read:user", "user:email"},
		Endpoint:     github.Endpoint,
	}
	logger.Infof("Configured GitHub auth for client %s", SecretsConfig.IdP.GitHub.Key)
}

func gitlabOauthConfigure() {
	if !SecretsConfig.IdP.GitLab.Usable() {
		logger.Debug("GitLab auth isn't configured or enabled")
		OAuthGitlabConfig = nil
	}

	OAuthGitlabConfig = &oauth2.Config{
		RedirectURL:  URLForAPI("oauth/gitlab/callback", nil),
		ClientID:     SecretsConfig.IdP.GitLab.Key,
		ClientSecret: SecretsConfig.IdP.GitLab.Secret,
		Scopes:       []string{"read_user"},
		Endpoint:     gitlab.Endpoint,
	}

	// Customise the endpoint, if a custom GitLab URL is given
	if CLIFlags.GitLabURL != "" {
		OAuthGitlabConfig.Endpoint.AuthURL = CLIFlags.GitLabURL + "/oauth/authorize"
		OAuthGitlabConfig.Endpoint.TokenURL = CLIFlags.GitLabURL + "/oauth/token"
	}
	logger.Infof("Configured GitLab auth for client %s", SecretsConfig.IdP.GitLab.Key)
}

func googleOauthConfigure() {
	if !SecretsConfig.IdP.Google.Usable() {
		logger.Debug("Google auth isn't configured or enabled")
		OAuthGoogleConfig = nil
	}

	OAuthGoogleConfig = &oauth2.Config{
		RedirectURL:  URLForAPI("oauth/google/callback", nil),
		ClientID:     SecretsConfig.IdP.Google.Key,
		ClientSecret: SecretsConfig.IdP.Google.Secret,
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.profile", "https://www.googleapis.com/auth/userinfo.email"},
		Endpoint:     google.Endpoint,
	}
	logger.Infof("Configured Google auth for client %s", SecretsConfig.IdP.Google.Key)
}
