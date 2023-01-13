package main

import (
	"fmt"
	"net/http"
)

func gitlabRedirectHandler(w http.ResponseWriter, r *http.Request) {
	if gitlabConfig == nil {
		logger.Errorf("gitlab oauth access attempt without configuration")
		_, _ = fmt.Fprintf(w, "error: this website has not configured gitlab OAuth")
		return
	}

	commenterToken := r.FormValue("commenterToken")

	_, err := commenterGetByCommenterToken(commenterToken)
	if err != nil && err != errorNoSuchToken {
		_, _ = fmt.Fprintf(w, "error: %s\n", err.Error())
		return
	}

	url := gitlabConfig.AuthCodeURL(commenterToken)
	http.Redirect(w, r, url, http.StatusFound)
}
