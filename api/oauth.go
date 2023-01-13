package main

var googleConfigured bool
var githubConfigured bool
var gitlabConfigured bool

func oauthConfigure() error {
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
