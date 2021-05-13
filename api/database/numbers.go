package database

import (
	"github.com/allinbits/demeris-backend/models"
	"github.com/jmoiron/sqlx"
)

func (d *Database) Numbers(address string) ([]models.AuthRow, error) {
	var numbers []models.AuthRow

	q, args, err := sqlx.In("SELECT * FROM tracelistener.auth WHERE address IN (?);", []string{address})
	if err != nil {
		return nil, err
	}

	q = d.dbi.DB.Rebind(q)

	return numbers, d.dbi.DB.Select(&numbers, q, args...)
}
