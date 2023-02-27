package e2e

// End2EndApp describes an application under e2e test
type End2EndApp interface {
	// DBExec executes the provided statement(s) against the database
	DBExec(query string, args ...any) error
	// DBInit initialises the database
	DBInit() error
	// LogInfo outputs a record of level info to the log
	LogInfo(fmt string, args ...any)
	// LogWarning outputs a record of level warning to the log
	LogWarning(fmt string, args ...any)
	// LogError outputs a record of level error to the log
	LogError(fmt string, args ...any)
}

// End2EndHandler describes an e2e testing plugin
type End2EndHandler interface {
	// Init binds the app under test to the plugin
	Init(app End2EndApp) error
	// HandleReset resets the backend to its initial state
	HandleReset() error
}
