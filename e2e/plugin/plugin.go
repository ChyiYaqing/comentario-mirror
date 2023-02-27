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

	// Install the seed
	if err := h.dbSeed(); err != nil {
		return err
	}

	h.app.LogInfo("Initialised e2e plugin")
	return nil
}

func (h *handler) HandleReset() error {
	h.app.LogInfo("Recreating the database schema")

	// Drop and recreate the public schema
	if err := h.app.DBExec("drop schema public cascade; create schema public;"); err != nil {
		return err
	}

	// Init the DB
	if err := h.app.DBInit(); err != nil {
		return err
	}

	// Install the seed
	return h.dbSeed()
}

// dbSeed installs seed data in the database
func (h *handler) dbSeed() error {
	h.app.LogInfo("Seeding the database")
	return h.app.DBExec(dbSeedSQL)
}
