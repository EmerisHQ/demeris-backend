package database

import "github.com/allinbits/demeris-backend/models"

func (d *Database) Pools() ([]models.PoolRow, error) {
	var pools []models.PoolRow

	q := "SELECT * FROM tracelistener.liquidity_pools;"

	q = d.dbi.DB.Rebind(q)

	return pools, d.dbi.DB.Select(&pools, q)
}
