package api

import (
	"gitlab.com/comentario/comentario/internal/svc"
	"gitlab.com/comentario/comentario/internal/util"
)

func commentStatistics(domain string) ([]int64, error) {
	statement := `
		select COUNT(comments.creationDate)
		from (
			select to_char(date_trunc('day', (current_date - offs)), 'YYYY-MM-DD') as date
			from generate_series(0, 30, 1) as offs
		) gen 
		    left outer join comments
			on 
				gen.date = to_char(date_trunc('day', comments.creationDate), 'YYYY-MM-DD') and
				comments.domain=$1
		group by gen.date
		order by gen.date;
	`
	rows, err := svc.DB.Query(statement, domain)
	if err != nil {
		logger.Errorf("cannot get daily views: %v", err)
		return []int64{}, util.ErrorInternal
	}

	defer rows.Close()

	last30Days := []int64{}
	for rows.Next() {
		var count int64
		if err = rows.Scan(&count); err != nil {
			logger.Errorf("cannot get daily comments for the last month: %v", err)
			return make([]int64, 0), util.ErrorInternal
		}
		last30Days = append(last30Days, count)
	}

	return last30Days, nil
}
