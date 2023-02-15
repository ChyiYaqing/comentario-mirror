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

	api.CommentApproveHandler = operations.CommentApproveHandlerFunc(func(params operations.CommentApproveParams) middleware.Responder {
		return middleware.NotImplemented("operation operations.CommentApprove has not yet been implemented")
	})
	api.CommentCountHandler = operations.CommentCountHandlerFunc(func(params operations.CommentCountParams) middleware.Responder {
		return middleware.NotImplemented("operation operations.CommentCount has not yet been implemented")
	})
	api.CommentDeleteHandler = operations.CommentDeleteHandlerFunc(func(params operations.CommentDeleteParams) middleware.Responder {
		return middleware.NotImplemented("operation operations.CommentDelete has not yet been implemented")
	})
	api.CommentEditHandler = operations.CommentEditHandlerFunc(func(params operations.CommentEditParams) middleware.Responder {
		return middleware.NotImplemented("operation operations.CommentEdit has not yet been implemented")
	})
	api.CommentListHandler = operations.CommentListHandlerFunc(func(params operations.CommentListParams) middleware.Responder {
		return middleware.NotImplemented("operation operations.CommentList has not yet been implemented")
	})
	api.CommentNewHandler = operations.CommentNewHandlerFunc(func(params operations.CommentNewParams) middleware.Responder {
		return middleware.NotImplemented("operation operations.CommentNew has not yet been implemented")
	})
	api.CommentVoteHandler = operations.CommentVoteHandlerFunc(func(params operations.CommentVoteParams) middleware.Responder {
		return middleware.NotImplemented("operation operations.CommentVote has not yet been implemented")
	})
	api.CommenterLoginHandler = operations.CommenterLoginHandlerFunc(func(params operations.CommenterLoginParams) middleware.Responder {
		return middleware.NotImplemented("operation operations.CommenterLogin has not yet been implemented")
	})
	api.CommenterNewHandler = operations.CommenterNewHandlerFunc(func(params operations.CommenterNewParams) middleware.Responder {
		return middleware.NotImplemented("operation operations.CommenterNew has not yet been implemented")
	})
	api.CommenterPhotoHandler = operations.CommenterPhotoHandlerFunc(func(params operations.CommenterPhotoParams) middleware.Responder {
		return middleware.NotImplemented("operation operations.CommenterPhoto has not yet been implemented")
	})
	api.CommenterSelfHandler = operations.CommenterSelfHandlerFunc(func(params operations.CommenterSelfParams) middleware.Responder {
		return middleware.NotImplemented("operation operations.CommenterSelf has not yet been implemented")
	})
	api.CommenterTokenNewHandler = operations.CommenterTokenNewHandlerFunc(func(params operations.CommenterTokenNewParams) middleware.Responder {
		return middleware.NotImplemented("operation operations.CommenterTokenNew has not yet been implemented")
	})
	api.CommenterUpdateHandler = operations.CommenterUpdateHandlerFunc(func(params operations.CommenterUpdateParams) middleware.Responder {
		return middleware.NotImplemented("operation operations.CommenterUpdate has not yet been implemented")
	})
	api.DomainClearHandler = operations.DomainClearHandlerFunc(func(params operations.DomainClearParams) middleware.Responder {
		return middleware.NotImplemented("operation operations.DomainClear has not yet been implemented")
	})
	api.DomainDeleteHandler = operations.DomainDeleteHandlerFunc(func(params operations.DomainDeleteParams) middleware.Responder {
		return middleware.NotImplemented("operation operations.DomainDelete has not yet been implemented")
	})
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
	api.DomainListHandler = operations.DomainListHandlerFunc(func(params operations.DomainListParams) middleware.Responder {
		return middleware.NotImplemented("operation operations.DomainList has not yet been implemented")
	})
	api.DomainModeratorDeleteHandler = operations.DomainModeratorDeleteHandlerFunc(func(params operations.DomainModeratorDeleteParams) middleware.Responder {
		return middleware.NotImplemented("operation operations.DomainModeratorDelete has not yet been implemented")
	})
	api.DomainModeratorNewHandler = operations.DomainModeratorNewHandlerFunc(func(params operations.DomainModeratorNewParams) middleware.Responder {
		return middleware.NotImplemented("operation operations.DomainModeratorNew has not yet been implemented")
	})
	api.DomainNewHandler = operations.DomainNewHandlerFunc(func(params operations.DomainNewParams) middleware.Responder {
		return middleware.NotImplemented("operation operations.DomainNew has not yet been implemented")
	})
	api.DomainSsoNewHandler = operations.DomainSsoNewHandlerFunc(func(params operations.DomainSsoNewParams) middleware.Responder {
		return middleware.NotImplemented("operation operations.DomainSsoNew has not yet been implemented")
	})
	api.DomainStatisticsHandler = operations.DomainStatisticsHandlerFunc(func(params operations.DomainStatisticsParams) middleware.Responder {
		return middleware.NotImplemented("operation operations.DomainStatistics has not yet been implemented")
	})
	api.DomainUpdateHandler = operations.DomainUpdateHandlerFunc(func(params operations.DomainUpdateParams) middleware.Responder {
		return middleware.NotImplemented("operation operations.DomainUpdate has not yet been implemented")
	})
	api.EmailGetHandler = operations.EmailGetHandlerFunc(func(params operations.EmailGetParams) middleware.Responder {
		return middleware.NotImplemented("operation operations.EmailGet has not yet been implemented")
	})
	api.EmailModerateHandler = operations.EmailModerateHandlerFunc(func(params operations.EmailModerateParams) middleware.Responder {
		return middleware.NotImplemented("operation operations.EmailModerate has not yet been implemented")
	})
	api.EmailUpdateHandler = operations.EmailUpdateHandlerFunc(func(params operations.EmailUpdateParams) middleware.Responder {
		return middleware.NotImplemented("operation operations.EmailUpdate has not yet been implemented")
	})
	api.ForgotHandler = operations.ForgotHandlerFunc(func(params operations.ForgotParams) middleware.Responder {
		return middleware.NotImplemented("operation operations.Forgot has not yet been implemented")
	})
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
	api.OwnerConfirmHexHandler = operations.OwnerConfirmHexHandlerFunc(func(params operations.OwnerConfirmHexParams) middleware.Responder {
		return middleware.NotImplemented("operation operations.OwnerConfirmHex has not yet been implemented")
	})
	api.OwnerDeleteHandler = operations.OwnerDeleteHandlerFunc(func(params operations.OwnerDeleteParams) middleware.Responder {
		return middleware.NotImplemented("operation operations.OwnerDelete has not yet been implemented")
	})
	api.OwnerLoginHandler = operations.OwnerLoginHandlerFunc(func(params operations.OwnerLoginParams) middleware.Responder {
		return middleware.NotImplemented("operation operations.OwnerLogin has not yet been implemented")
	})
	api.OwnerNewHandler = operations.OwnerNewHandlerFunc(func(params operations.OwnerNewParams) middleware.Responder {
		return middleware.NotImplemented("operation operations.OwnerNew has not yet been implemented")
	})
	api.OwnerSelfHandler = operations.OwnerSelfHandlerFunc(handlers.OwnerSelf)
	// Page
	api.PageUpdateHandler = operations.PageUpdateHandlerFunc(func(params operations.PageUpdateParams) middleware.Responder {
		return middleware.NotImplemented("operation operations.PageUpdate has not yet been implemented")
	})
	api.ResetHandler = operations.ResetHandlerFunc(func(params operations.ResetParams) middleware.Responder {
		return middleware.NotImplemented("operation operations.Reset has not yet been implemented")
	})

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
	exitIfError(api.OAuthConfigure())
	exitIfError(api.MarkdownRendererCreate())
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
