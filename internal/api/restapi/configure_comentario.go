// This file is safe to edit. Once it exists it will not be overwritten

package restapi

import (
	"crypto/tls"
	"fmt"
	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/justinas/alice"
	"github.com/op/go-logging"
	"gitlab.com/comentario/comentario/internal/api"
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
	api.DomainExportBeginHandler = operations.DomainExportBeginHandlerFunc(func(params operations.DomainExportBeginParams) middleware.Responder {
		return middleware.NotImplemented("operation operations.DomainExportBegin has not yet been implemented")
	})
	api.DomainExportDownloadHandler = operations.DomainExportDownloadHandlerFunc(func(params operations.DomainExportDownloadParams) middleware.Responder {
		return middleware.NotImplemented("operation operations.DomainExportDownload has not yet been implemented")
	})
	api.DomainImportCommentoHandler = operations.DomainImportCommentoHandlerFunc(func(params operations.DomainImportCommentoParams) middleware.Responder {
		return middleware.NotImplemented("operation operations.DomainImportCommento has not yet been implemented")
	})
	api.DomainImportDisqusHandler = operations.DomainImportDisqusHandlerFunc(func(params operations.DomainImportDisqusParams) middleware.Responder {
		return middleware.NotImplemented("operation operations.DomainImportDisqus has not yet been implemented")
	})
	api.DomainListHandler = operations.DomainListHandlerFunc(handlers.DomainList)
	api.DomainModeratorDeleteHandler = operations.DomainModeratorDeleteHandlerFunc(func(params operations.DomainModeratorDeleteParams) middleware.Responder {
		return middleware.NotImplemented("operation operations.DomainModeratorDelete has not yet been implemented")
	})
	api.DomainModeratorNewHandler = operations.DomainModeratorNewHandlerFunc(handlers.DomainModeratorNew)
	api.DomainNewHandler = operations.DomainNewHandlerFunc(handlers.DomainNew)
	api.DomainSsoNewHandler = operations.DomainSsoNewHandlerFunc(func(params operations.DomainSsoNewParams) middleware.Responder {
		return middleware.NotImplemented("operation operations.DomainSsoNew has not yet been implemented")
	})
	api.DomainStatisticsHandler = operations.DomainStatisticsHandlerFunc(func(params operations.DomainStatisticsParams) middleware.Responder {
		return middleware.NotImplemented("operation operations.DomainStatistics has not yet been implemented")
	})
	api.DomainUpdateHandler = operations.DomainUpdateHandlerFunc(handlers.DomainUpdate)
	// Email
	api.EmailGetHandler = operations.EmailGetHandlerFunc(handlers.EmailGet)
	api.EmailModerateHandler = operations.EmailModerateHandlerFunc(func(params operations.EmailModerateParams) middleware.Responder {
		return middleware.NotImplemented("operation operations.EmailModerate has not yet been implemented")
	})
	api.EmailUpdateHandler = operations.EmailUpdateHandlerFunc(func(params operations.EmailUpdateParams) middleware.Responder {
		return middleware.NotImplemented("operation operations.EmailUpdate has not yet been implemented")
	})
	// OAuth
	api.OauthGithubCallbackHandler = operations.OauthGithubCallbackHandlerFunc(func(params operations.OauthGithubCallbackParams) middleware.Responder {
		return middleware.NotImplemented("operation operations.OauthGithubCallback has not yet been implemented")
	})
	api.OauthGithubRedirectHandler = operations.OauthGithubRedirectHandlerFunc(func(params operations.OauthGithubRedirectParams) middleware.Responder {
		return middleware.NotImplemented("operation operations.OauthGithubRedirect has not yet been implemented")
	})
	api.OauthGitlabCallbackHandler = operations.OauthGitlabCallbackHandlerFunc(func(params operations.OauthGitlabCallbackParams) middleware.Responder {
		return middleware.NotImplemented("operation operations.OauthGitlabCallback has not yet been implemented")
	})
	api.OauthGitlabRedirectHandler = operations.OauthGitlabRedirectHandlerFunc(func(params operations.OauthGitlabRedirectParams) middleware.Responder {
		return middleware.NotImplemented("operation operations.OauthGitlabRedirect has not yet been implemented")
	})
	api.OauthGoogleCallbackHandler = operations.OauthGoogleCallbackHandlerFunc(func(params operations.OauthGoogleCallbackParams) middleware.Responder {
		return middleware.NotImplemented("operation operations.OauthGoogleCallback has not yet been implemented")
	})
	api.OauthGoogleRedirectHandler = operations.OauthGoogleRedirectHandlerFunc(func(params operations.OauthGoogleRedirectParams) middleware.Responder {
		return middleware.NotImplemented("operation operations.OauthGoogleRedirect has not yet been implemented")
	})
	api.OauthSsoCallbackHandler = operations.OauthSsoCallbackHandlerFunc(func(params operations.OauthSsoCallbackParams) middleware.Responder {
		return middleware.NotImplemented("operation operations.OauthSsoCallback has not yet been implemented")
	})
	api.OauthSsoRedirectHandler = operations.OauthSsoRedirectHandlerFunc(func(params operations.OauthSsoRedirectParams) middleware.Responder {
		return middleware.NotImplemented("operation operations.OauthSsoRedirect has not yet been implemented")
	})
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
	exitIfError(api.DomainExportCleanupBegin())
	exitIfError(api.ViewsCleanupBegin())
	exitIfError(api.SSOTokenCleanupBegin())
	//TODO replaced with OpenAPI exitIfError(api.RoutesServe())

}

func exitIfError(err error) {
	if err != nil {
		fmt.Printf("fatal error: %v\n", err)
		os.Exit(1)
	}
}
