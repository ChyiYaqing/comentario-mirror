package api

var googleConfigured bool
var githubConfigured bool
var gitlabConfigured bool

func OAuthConfigure() error {
	if err := googleOauthConfigure(); err != nil {
		return err
	}

	if err := githubOauthConfigure(); err != nil {
		return err
	}

	if err := gitlabOauthConfigure(); err != nil {
		return err
	}

	return nil
}
