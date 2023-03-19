package main

//go:generate rm -rf internal/api/models internal/api/restapi/operations
//go:generate swagger generate server --exclude-main --name Comentario --target internal/api --spec ./swagger/swagger.yml --principal gitlab.com/comentario/comentario/internal/data.Principal --principal-is-interface

import (
	"fmt"
	"github.com/go-openapi/loads"
	"github.com/jessevdk/go-flags"
	"github.com/op/go-logging"
	"gitlab.com/comentario/comentario/internal/api/restapi"
	"gitlab.com/comentario/comentario/internal/api/restapi/operations"
	"gitlab.com/comentario/comentario/internal/config"
	"gitlab.com/comentario/comentario/internal/util"
	"os"
)

// logger represents a package-wide logger instance
var logger = logging.MustGetLogger("main")

// Variables set during build
var (
	version = "(?)"
	date    = "(?)"
)

func main() {
	// Load the embedded Swagger file
	swaggerSpec, err := loads.Analyzed(restapi.SwaggerJSON, "")
	if err != nil {
		logger.Fatal(err)
	}

	// Create new service API
	apiInstance := operations.NewComentarioAPI(swaggerSpec)
	server := restapi.NewServer(apiInstance)
	defer server.Shutdown()

	// Configure command-line options
	parser := flags.NewParser(server, flags.Default)
	parser.ShortDescription = util.ApplicationName
	parser.LongDescription = swaggerSpec.Spec().Info.Description
	server.ConfigureFlags()
	for _, optsGroup := range apiInstance.CommandLineOptionsGroups {
		_, err := parser.AddGroup(optsGroup.ShortDescription, optsGroup.LongDescription, optsGroup.Options)
		if err != nil {
			logger.Fatal(err)
		}
	}

	// Parse the command line
	if _, err := parser.Parse(); err != nil {
		code := 1
		if fe, ok := err.(*flags.Error); ok {
			if fe.Type == flags.ErrHelp {
				code = 0
			}
		}
		os.Exit(code)
	}

	// Configure variables
	config.AppVersion = version
	config.BuildDate = date
	fmt.Printf("Comentario server, version %s, built %s\n", config.AppVersion, config.BuildDate)

	// Configure logging
	var logLevel logging.Level
	switch len(config.CLIFlags.Verbose) {
	case 0:
		logLevel = logging.WARNING
	case 1:
		logLevel = logging.INFO
	default:
		logLevel = logging.DEBUG
	}
	logging.SetFormatter(logging.MustStringFormatter(`%{level:-5s} %{module} %{message}`))
	logging.SetLevel(logLevel, "")

	// Configure the API
	server.ConfigureAPI()

	// serve API
	if err := server.Serve(); err != nil {
		logger.Fatalf("Serve() failed: %v", err)
	}
}
