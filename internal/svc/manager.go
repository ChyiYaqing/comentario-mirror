package svc

import (
	"gitlab.com/comentario/comentario/internal/persistence"
)

// TheServiceManager is a global service manager interface
var TheServiceManager ServiceManager = &manager{}

// obsolete
var DB *persistence.Database

// db is a global database instance (only available for the services package)
var db *persistence.Database

// ServiceManager provides high-level service management routines
type ServiceManager interface {
	// Initialise performs necessary initialisation of the services
	Initialise()
	// Shutdown performs necessary teardown of the services
	Shutdown()
}

//----------------------------------------------------------------------------------------------------------------------

type manager struct {
	inited bool
}

func (m *manager) Initialise() {
	if m.inited {
		logger.Fatal("ServiceManager is already initialised")
	}
	m.inited = true

	// Initiate a DB connection
	var err error
	if db, err = persistence.InitDB(); err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	DB = db // TODO REMOVE

	// Start the cleanup service
	if err = TheCleanupService.Init(); err != nil {
		logger.Fatalf("Failed to initialise cleanup service: %v", err)
	}

	// Start the version service
	TheVersionCheckService.Init()
}

func (m *manager) Shutdown() {
	// Make sure the services are initialised
	if !m.inited {
		return
	}

	// Teardown the database
	_ = db.Shutdown()
	db = nil
	DB = db // TODO REMOVE
	m.inited = false
}
