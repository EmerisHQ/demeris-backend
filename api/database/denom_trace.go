package database

import (
	"github.com/allinbits/demeris-backend/models"
)

func (d *Database) DenomTrace(chain string, hash string) (models.DenomTrace, error) {
	var denomTraces models.DenomTrace

	q := "SELECT * FROM tracelistener.denom_traces WHERE chain_name=? and hash=? limit 1;"

	q = d.dbi.DB.Rebind(q)

	return denomTraces, d.dbi.DB.Select(&denomTraces, q, chain, hash)
}
