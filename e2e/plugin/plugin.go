package main

import "gitlab.com/comentario/comentario/internal/e2e"
import _ "embed"

// Handler is the exported plugin implementation
//
//goland:noinspection GoUnusedGlobalVariable
var Handler e2e.End2EndHandler = &handler{}

//go:embed db-seed.sql
var dbSeedSQL string

// handler is an End2EndHandler implementation
type handler struct {
	app e2e.End2EndApp // Host app
}

func (h *handler) Init(app e2e.End2EndApp) error {
	h.app = app

	// Reinit the DB to install the seed
	if err := h.app.RecreateDBSchema(dbSeedSQL); err != nil {
		return err
	}

	h.app.LogInfo("Initialised e2e plugin")
	return nil
}

func (h *handler) HandleReset() error {
	h.app.LogInfo("Recreating the database schema")

	// Drop and recreate the public schema
	return h.app.RecreateDBSchema(dbSeedSQL)
}
