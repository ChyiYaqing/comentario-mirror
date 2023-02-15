package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-openapi/strfmt"
	"gitlab.com/comentario/comentario/internal/util"
	"io"
	"net/http"
	"os"
)

func gitlabCallbackHandler(w http.ResponseWriter, r *http.Request) {
	commenterToken := r.FormValue("state")
	code := r.FormValue("code")

	_, err := commenterGetByCommenterToken(commenterToken)
	if err != nil && err != util.ErrorNoSuchToken {
		_, _ = fmt.Fprintf(w, "Error: %s\n", err.Error())
		return
	}

	token, err := gitlabConfig.Exchange(context.TODO(), code)
	if err != nil {
		_, _ = fmt.Fprintf(w, "Error: %s", err.Error())
		return
	}

	resp, err := http.Get(os.Getenv("GITLAB_URL") + "/api/v4/user?access_token=" + token.AccessToken)
	if err != nil {
		_, _ = fmt.Fprintf(w, "Error: %s", err.Error())
		return
	}
	logger.Infof("%v", resp.StatusCode)
	defer resp.Body.Close()

	contents, err := io.ReadAll(resp.Body)
	if err != nil {
		_, _ = fmt.Fprintf(w, "Error: %s", util.ErrorCannotReadResponse.Error())
		return
	}

	user := make(map[string]interface{})
	if err := json.Unmarshal(contents, &user); err != nil {
		_, _ = fmt.Fprintf(w, "Error: %s", util.ErrorInternal.Error())
		return
	}

	if user["email"] == nil {
		_, _ = fmt.Fprintf(w, "Error: no email address returned by Gitlab")
		return
	}

	email := user["email"].(string)

	if user["name"] == nil {
		_, _ = fmt.Fprintf(w, "Error: no name returned by Gitlab")
		return
	}

	name := user["name"].(string)

	link := "undefined"
	if user["web_url"] != nil {
		link = user["web_url"].(string)
	}

	photo := "undefined"
	if user["avatar_url"] != nil {
		photo = user["avatar_url"].(string)
	}

	c, err := commenterGetByEmail("gitlab", strfmt.Email(email))
	if err != nil && err != util.ErrorNoSuchCommenter {
		_, _ = fmt.Fprintf(w, "Error: %s", err.Error())
		return
	}

	var commenterHex string

	if err == util.ErrorNoSuchCommenter {
		commenterHex, err = commenterNew(strfmt.Email(email), name, link, photo, "gitlab", "")
		if err != nil {
			_, _ = fmt.Fprintf(w, "Error: %s", err.Error())
			return
		}
	} else {
		if err = commenterUpdate(c.CommenterHex, email, name, link, photo, "gitlab"); err != nil {
			logger.Warningf("cannot update commenter: %s", err)
			// not a serious enough to exit with an error
		}

		commenterHex = c.CommenterHex
	}

	if err := commenterSessionUpdate(commenterToken, commenterHex); err != nil {
		_, _ = fmt.Fprintf(w, "Error: %s", err.Error())
		return
	}

	_, _ = fmt.Fprintf(w, "<html><script>window.parent.close()</script></html>")
}
