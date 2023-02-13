package api

import "gitlab.com/comentario/comentario/internal/util"

func pageTitleUpdate(domain string, path string) (string, error) {
	title, err := util.HTMLTitleGet("http://" + domain + path)
	if err != nil {
		// This could fail due to a variety of reasons that we can't control such
		// as the user's URL 404 or something, so let's not pollute the error log
		// with messages. Just use a sane title. Maybe we'll have the ability to
		// retry later.
		logger.Errorf("%v", err)
		title = domain
	}

	statement := `update pages set title = $3 where domain = $1 and path = $2;`
	_, err = DB.Exec(statement, domain, path, title)
	if err != nil {
		logger.Errorf("cannot update pages table with title: %v", err)
		return "", err
	}

	return title, nil
}
