package api

import (
	"fmt"
	"gitlab.com/comentario/comentario/internal/util"
	"net/http"
)

func githubRedirectHandler(w http.ResponseWriter, r *http.Request) {
	if githubConfig == nil {
		logger.Errorf("github oauth access attempt without configuration")
		_, _ = fmt.Fprintf(w, "error: this website has not configured github OAuth")
		return
	}

	commenterToken := r.FormValue("commenterToken")

	_, err := commenterGetByCommenterToken(commenterToken)
	if err != nil && err != util.ErrorNoSuchToken {
		_, _ = fmt.Fprintf(w, "error: %s\n", err.Error())
		return
	}

	url := githubConfig.AuthCodeURL(commenterToken)
	http.Redirect(w, r, url, http.StatusFound)
}
