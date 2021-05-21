package database

import (
	"github.com/allinbits/demeris-backend/models"
)

func (d *Database) Balances(address string) ([]models.BalanceRow, error) {
	var balances []models.BalanceRow

	q := "SELECT * FROM tracelistener.balances WHERE address=? and chain_name in (select chain_name from cns.chains where enabled=true);"

	q = d.dbi.DB.Rebind(q)

	return balances, d.dbi.DB.Select(&balances, q, address)
}
