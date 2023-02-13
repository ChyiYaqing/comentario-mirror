package api

import (
	"fmt"
	"gitlab.com/comentario/comentario/internal/svc"
	"net/http"
	"time"
)

func domainExportDownloadHandler(w http.ResponseWriter, r *http.Request) {
	exportHex := r.FormValue("exportHex")
	if exportHex == "" {
		_, _ = fmt.Fprintf(w, "Error: empty exportHex\n")
		return
	}

	statement := `select domain, binData, creationDate from exports where exportHex = $1;`
	row := svc.DB.QueryRow(statement, exportHex)

	var domain string
	var binData []byte
	var creationDate time.Time
	if err := row.Scan(&domain, &binData, &creationDate); err != nil {
		_, _ = fmt.Fprintf(w, "Error: that exportHex does not exist\n")
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s-%v.json.gz"`, domain, creationDate.Unix()))
	_, _ = w.Write(binData)
}
