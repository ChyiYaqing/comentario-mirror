// This file is safe to edit. Once it exists it will not be overwritten

package restapi

import (
	"crypto/tls"
	"fmt"
	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/justinas/alice"
	"github.com/op/go-logging"
	"gitlab.com/comentario/comentario/internal/api/restapi/handlers"
	"gitlab.com/comentario/comentario/internal/api/restapi/operations"
	"gitlab.com/comentario/comentario/internal/config"
	"gitlab.com/comentario/comentario/internal/mail"
	"gitlab.com/comentario/comentario/internal/svc"
	"gitlab.com/comentario/comentario/internal/util"
	"net/http"
	"os"
)

// logger represents a package-wide logger instance
var logger = logging.MustGetLogger("restapi")

func configureFlags(api *operations.ComentarioAPI) {
	api.CommandLineOptionsGroups = []swag.CommandLineOptionsGroup{
		{
			ShortDescription: "Server options",
			LongDescription:  "Server options",
			Options:          &config.CLIFlags,
		},
	}
}

func configureAPI(api *operations.ComentarioAPI) http.Handler {
	api.ServeError = errors.ServeError
	api.Logger = logger.Infof
	api.JSONConsumer = runtime.JSONConsumer()
	api.JSONProducer = runtime.JSONProducer()
	api.UrlformConsumer = runtime.DiscardConsumer

	// Use a more strict email validator than the default, RFC5322-compliant one
	eml := strfmt.Email("")
	api.Formats().Add("email", &eml, util.IsValidEmail)

	// Update the config based on the CLI flags
	if err := config.CLIParsed(); err != nil {
		logger.Fatalf("Failed to process configuration: %v", err)
	}

	// Configure swagger UI
	if config.CLIFlags.EnableSwaggerUI {
		api.UseSwaggerUI()
	}

	// Comment
	api.CommentApproveHandler = operations.CommentApproveHandlerFunc(handlers.CommentApprove)
	api.CommentCountHandler = operations.CommentCountHandlerFunc(handlers.CommentCount)
	api.CommentDeleteHandler = operations.CommentDeleteHandlerFunc(handlers.CommentDelete)
	api.CommentEditHandler = operations.CommentEditHandlerFunc(handlers.CommentEdit)
	api.CommentListHandler = operations.CommentListHandlerFunc(handlers.CommentList)
	api.CommentNewHandler = operations.CommentNewHandlerFunc(handlers.CommentNew)
	api.CommentVoteHandler = operations.CommentVoteHandlerFunc(handlers.CommentVote)
	// Commenter
	api.CommenterLoginHandler = operations.CommenterLoginHandlerFunc(handlers.CommenterLogin)
	api.CommenterNewHandler = operations.CommenterNewHandlerFunc(handlers.CommenterNew)
	api.CommenterPhotoHandler = operations.CommenterPhotoHandlerFunc(handlers.CommenterPhoto)
	api.CommenterSelfHandler = operations.CommenterSelfHandlerFunc(handlers.CommenterSelf)
	api.CommenterTokenNewHandler = operations.CommenterTokenNewHandlerFunc(handlers.CommenterTokenNew)
	api.CommenterUpdateHandler = operations.CommenterUpdateHandlerFunc(handlers.CommenterUpdate)
	// Domain
	api.DomainClearHandler = operations.DomainClearHandlerFunc(handlers.DomainClear)
	api.DomainDeleteHandler = operations.DomainDeleteHandlerFunc(handlers.DomainDelete)
	api.DomainExportBeginHandler = operations.DomainExportBeginHandlerFunc(handlers.DomainExportBegin)
	api.DomainExportDownloadHandler = operations.DomainExportDownloadHandlerFunc(handlers.DomainExportDownload)
	api.DomainImportCommentoHandler = operations.DomainImportCommentoHandlerFunc(handlers.DomainImportCommento)
	api.DomainImportDisqusHandler = operations.DomainImportDisqusHandlerFunc(handlers.DomainImportDisqus)
	api.DomainListHandler = operations.DomainListHandlerFunc(handlers.DomainList)
	api.DomainModeratorDeleteHandler = operations.DomainModeratorDeleteHandlerFunc(handlers.DomainModeratorDelete)
	api.DomainModeratorNewHandler = operations.DomainModeratorNewHandlerFunc(handlers.DomainModeratorNew)
	api.DomainNewHandler = operations.DomainNewHandlerFunc(handlers.DomainNew)
	api.DomainSsoSecretNewHandler = operations.DomainSsoSecretNewHandlerFunc(handlers.DomainSsoSecretNew)
	api.DomainStatisticsHandler = operations.DomainStatisticsHandlerFunc(handlers.DomainStatistics)
	api.DomainUpdateHandler = operations.DomainUpdateHandlerFunc(handlers.DomainUpdate)
	// Email
	api.EmailGetHandler = operations.EmailGetHandlerFunc(handlers.EmailGet)
	api.EmailModerateHandler = operations.EmailModerateHandlerFunc(handlers.EmailModerate)
	api.EmailUpdateHandler = operations.EmailUpdateHandlerFunc(handlers.EmailUpdate)
	// OAuth
	api.OauthGithubCallbackHandler = operations.OauthGithubCallbackHandlerFunc(handlers.OauthGithubCallback)
	api.OauthGithubRedirectHandler = operations.OauthGithubRedirectHandlerFunc(handlers.OauthGithubRedirect)
	api.OauthGitlabCallbackHandler = operations.OauthGitlabCallbackHandlerFunc(handlers.OauthGitlabCallback)
	api.OauthGitlabRedirectHandler = operations.OauthGitlabRedirectHandlerFunc(handlers.OauthGitlabRedirect)
	api.OauthGoogleCallbackHandler = operations.OauthGoogleCallbackHandlerFunc(handlers.OauthGoogleCallback)
	api.OauthGoogleRedirectHandler = operations.OauthGoogleRedirectHandlerFunc(handlers.OauthGoogleRedirect)
	api.OauthSsoCallbackHandler = operations.OauthSsoCallbackHandlerFunc(handlers.OauthSsoCallback)
	api.OauthSsoRedirectHandler = operations.OauthSsoRedirectHandlerFunc(handlers.OauthSsoRedirect)
	// Owner
	api.OwnerConfirmHexHandler = operations.OwnerConfirmHexHandlerFunc(handlers.OwnerConfirmHex)
	api.OwnerDeleteHandler = operations.OwnerDeleteHandlerFunc(handlers.OwnerDelete)
	api.OwnerLoginHandler = operations.OwnerLoginHandlerFunc(handlers.OwnerLogin)
	api.OwnerNewHandler = operations.OwnerNewHandlerFunc(handlers.OwnerNew)
	api.OwnerSelfHandler = operations.OwnerSelfHandlerFunc(handlers.OwnerSelf)
	// Page
	api.PageUpdateHandler = operations.PageUpdateHandlerFunc(handlers.PageUpdate)
	// Auth
	api.ForgotPasswordHandler = operations.ForgotPasswordHandlerFunc(handlers.ForgotPassword)
	api.ResetPasswordHandler = operations.ResetPasswordHandlerFunc(handlers.ResetPassword)

	// Shutdown functions
	api.PreServerShutdown = func() {}
	api.ServerShutdown = svc.TheServiceManager.Shutdown

	// Set up the middleware
	chain := alice.New(
		rootRedirectHandler,
		corsHandler,
		staticHandler,
		makeAPIHandler(api.Serve(nil)),
	)

	// Finally add the fallback handlers
	return chain.Then(fallbackHandler())
}

// The TLS configuration before HTTPS server starts.
func configureTLS(_ *tls.Config) {
	// Not implemented
}

// configureServer is a callback that is invoked before the server startup with the protocol it's supposed to be
// handling (http, https, or unix)
func configureServer(_ *http.Server, scheme, _ string) {
	if scheme != "http" {
		return
	}

	// Initialise the services
	svc.TheServiceManager.Initialise()

	// TODO refactor all below
	exitIfError(config.ConfigParse())
	exitIfError(mail.SMTPConfigure())
	exitIfError(mail.SMTPTemplatesLoad())
	exitIfError(handlers.OAuthConfigure())
	exitIfError(config.VersionCheckStart())
	//TODO replaced with OpenAPI exitIfError(api.RoutesServe())

}

func exitIfError(err error) {
	if err != nil {
		fmt.Printf("fatal error: %v\n", err)
		os.Exit(1)
	}
}
