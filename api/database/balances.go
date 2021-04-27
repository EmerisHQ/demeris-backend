package database

import (
	"github.com/allinbits/demeris-backend/models"
)

func (d *Database) Balances(address string) ([]models.Balance, error) {
	var balances []models.Balance

	q := "SELECT * FROM tracelistener.balances WHERE address=?;"

	q = d.dbi.DB.Rebind(q)

	return balances, d.dbi.DB.Select(&balances, q, address)
}
