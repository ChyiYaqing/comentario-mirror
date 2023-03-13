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
	"gitlab.com/comentario/comentario/internal/api/restapi/handlers"
	"gitlab.com/comentario/comentario/internal/api/restapi/operations"
	"gitlab.com/comentario/comentario/internal/config"
	"gitlab.com/comentario/comentario/internal/e2e"
	"gitlab.com/comentario/comentario/internal/svc"
	"gitlab.com/comentario/comentario/internal/util"
	"net/http"
	"os"
	"path"
	"plugin"
)

// logger represents a package-wide logger instance
var logger = logging.MustGetLogger("restapi")

// Global e2e handler instance (only in e2e testing mode)
var e2eHandler e2e.End2EndHandler

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
	api.GzipProducer = runtime.ByteStreamProducer()
	api.HTMLProducer = runtime.TextProducer()
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
		logger.Warningf("Enabling Swagger UI")
		api.UseSwaggerUI()
	}

	// Set up auth handlers
	api.OwnerCookieAuth = FindOwnerByCookieHeader

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
	api.OauthInitHandler = operations.OauthInitHandlerFunc(handlers.OauthInit)
	api.OauthCallbackHandler = operations.OauthCallbackHandlerFunc(handlers.OauthCallback)
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

	// If in e2e-testing mode, configure the backend accordingly
	if config.CLIFlags.E2e {
		if err := configureE2eMode(api); err != nil {
			logger.Fatalf("Failed to configure e2e plugin: %v", err)
		}
	}

	// Set up the middleware
	chain := alice.New(
		redirectToLangRootHandler,
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

	// Init the e2e handler, if in the e2e testing mode
	if e2eHandler != nil {
		if err := e2eHandler.Init(&e2eApp{logger: logging.MustGetLogger("e2e")}); err != nil {
			logger.Fatalf("e2e handler init failed: %v", err)
		}
	}
}

// configureE2eMode configures the app to run in the end-2-end testing mode
func configureE2eMode(api *operations.ComentarioAPI) error {
	// Get the plugin path
	p, err := os.Executable()
	if err != nil {
		return err
	}
	pluginFile := path.Join(path.Dir(p), "comentario-e2e.so")

	// Load the e2e plugin
	logger.Infof("Loading e2e plugin %s", pluginFile)
	plug, err := plugin.Open(pluginFile)
	if err != nil {
		return err
	}

	// Look up the handler
	h, err := plug.Lookup("Handler")
	if err != nil {
		return err
	}

	// Fetch the service interface (hPtr is a pointer, because Lookup always returns a pointer to symbol)
	hPtr, ok := h.(*e2e.End2EndHandler)
	if !ok {
		return fmt.Errorf("symbol Handler from plugin %s doesn't implement End2EndHandler", pluginFile)
	}

	// Configure API endpoints
	e2eHandler = *hPtr
	api.E2eResetHandler = operations.E2eResetHandlerFunc(func(operations.E2eResetParams) middleware.Responder {
		if err := e2eHandler.HandleReset(); err != nil {
			logger.Errorf("E2eReset failed: %v", err)
			return operations.NewGenericInternalServerError()
		}
		return operations.NewE2eResetNoContent()
	})

	// Succeeded
	return nil
}
