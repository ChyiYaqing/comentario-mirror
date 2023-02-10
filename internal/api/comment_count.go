package api

import (
	"github.com/lib/pq"
	"gitlab.com/commento/commento/api/internal/util"
	"net/http"
)

func commentCount(domain string, paths []string) (map[string]int, error) {
	commentCounts := map[string]int{}

	if domain == "" {
		return nil, util.ErrorMissingField
	}

	if len(paths) == 0 {
		return nil, util.ErrorEmptyPaths
	}

	statement := `select path, commentCount from pages where domain = $1 and path = any($2);`
	rows, err := DB.Query(statement, domain, pq.Array(paths))
	if err != nil {
		logger.Errorf("cannot get comments: %v", err)
		return nil, util.ErrorInternal
	}
	defer rows.Close()

	for rows.Next() {
		var path string
		var commentCount int
		if err = rows.Scan(&path, &commentCount); err != nil {
			logger.Errorf("cannot scan path and commentCount: %v", err)
			return nil, util.ErrorInternal
		}

		commentCounts[path] = commentCount
	}

	return commentCounts, nil
}

func commentCountHandler(w http.ResponseWriter, r *http.Request) {
	type request struct {
		Domain *string   `json:"domain"`
		Paths  *[]string `json:"paths"`
	}

	var x request
	if err := BodyUnmarshal(r, &x); err != nil {
		BodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	domain := domainStrip(*x.Domain)

	commentCounts, err := commentCount(domain, *x.Paths)
	if err != nil {
		BodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}
	BodyMarshalChecked(w, response{"success": true, "commentCounts": commentCounts})
}
