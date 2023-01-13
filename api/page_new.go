package main

func pageNew(domain string, path string) error {
	// path can be empty
	if domain == "" {
		return errorMissingField
	}

	statement := `insert into pages(domain, path) values($1, $2) on CONFLICT DO NOTHING;`
	_, err := db.Exec(statement, domain, path)
	if err != nil {
		logger.Errorf("error inserting new page: %v", err)
		return errorInternal
	}

	return nil
}
