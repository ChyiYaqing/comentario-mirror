package api

import (
	"gitlab.com/commento/commento/api/internal/util"
	"time"
)

type moderator struct {
	Email   string    `json:"email"`
	Domain  string    `json:"domain"`
	AddDate time.Time `json:"addDate"`
}

func domainModeratorList(domain string) ([]moderator, error) {
	statement := `
		select email, addDate
		from moderators
		where domain=$1;
	`
	rows, err := DB.Query(statement, domain)
	if err != nil {
		logger.Errorf("cannot get moderators: %v", err)
		return nil, util.ErrorInternal
	}
	defer rows.Close()

	moderators := []moderator{}
	for rows.Next() {
		m := moderator{}
		if err = rows.Scan(&m.Email, &m.AddDate); err != nil {
			logger.Errorf("cannot Scan moderator: %v", err)
			return nil, util.ErrorInternal
		}

		moderators = append(moderators, m)
	}

	return moderators, nil
}

func isDomainModerator(domain string, email string) (bool, error) {
	statement := `
		select EXISTS (
			select 1
			from moderators
			where domain=$1 and email=$2
		);
	`
	row := DB.QueryRow(statement, domain, email)

	var exists bool
	if err := row.Scan(&exists); err != nil {
		logger.Errorf("cannot query if moderator: %v", err)
		return false, util.ErrorInternal
	}

	return exists, nil
}
