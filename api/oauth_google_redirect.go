package main

import (
	"fmt"
	"net/http"
)

func googleRedirectHandler(w http.ResponseWriter, r *http.Request) {
	if googleConfig == nil {
		logger.Errorf("google oauth access attempt without configuration")
		_, _ = fmt.Fprintf(w, "error: this website has not configured Google OAuth")
		return
	}

	commenterToken := r.FormValue("commenterToken")

	_, err := commenterGetByCommenterToken(commenterToken)
	if err != nil && err != errorNoSuchToken {
		_, _ = fmt.Fprintf(w, "error: %s\n", err.Error())
		return
	}

	url := googleConfig.AuthCodeURL(commenterToken)
	http.Redirect(w, r, url, http.StatusFound)
}
