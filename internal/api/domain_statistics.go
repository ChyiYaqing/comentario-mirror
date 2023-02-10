package api

import (
	"gitlab.com/commento/commento/api/internal/util"
	"net/http"
)

func domainStatistics(domain string) ([]int64, error) {
	statement := `
		select COUNT(views.viewDate)
		from (
			select to_char(date_trunc('day', (current_date - offs)), 'YYYY-MM-DD') as date
			from generate_series(0, 30, 1) as offs
		) gen left outer join views
		on gen.date = to_char(date_trunc('day', views.viewDate), 'YYYY-MM-DD') and
		   views.domain=$1
		group by gen.date
		order by gen.date;
	`
	rows, err := DB.Query(statement, domain)
	if err != nil {
		logger.Errorf("cannot get daily views: %v", err)
		return []int64{}, util.ErrorInternal
	}

	defer rows.Close()

	last30Days := []int64{}
	for rows.Next() {
		var count int64
		if err = rows.Scan(&count); err != nil {
			logger.Errorf("cannot get daily views for the last month: %v", err)
			return []int64{}, util.ErrorInternal
		}
		last30Days = append(last30Days, count)
	}

	return last30Days, nil
}

func domainStatisticsHandler(w http.ResponseWriter, r *http.Request) {
	type request struct {
		OwnerToken *string `json:"ownerToken"`
		Domain     *string `json:"domain"`
	}

	var x request
	if err := BodyUnmarshal(r, &x); err != nil {
		BodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	o, err := ownerGetByOwnerToken(*x.OwnerToken)
	if err != nil {
		BodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	domain := domainStrip(*x.Domain)
	isOwner, err := domainOwnershipVerify(o.OwnerHex, domain)
	if err != nil {
		BodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	if !isOwner {
		BodyMarshalChecked(w, response{"success": false, "message": util.ErrorNotAuthorised.Error()})
		return
	}

	viewsLast30Days, err := domainStatistics(domain)
	if err != nil {
		BodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	commentsLast30Days, err := commentStatistics(domain)
	if err != nil {
		BodyMarshalChecked(w, response{"success": false, "message": err.Error()})
		return
	}

	BodyMarshalChecked(w, response{"success": true, "viewsLast30Days": viewsLast30Days, "commentsLast30Days": commentsLast30Days})
}
