package database

import "github.com/allinbits/demeris-backend/models"

func (d *Database) Connection(chain string, connection_id string) (models.Client, error) {
	var client models.Client

	q := `
	SELECT *
	FROM tracelistener.connections 
	WHERE chain_name=? AND connection_id=?;
	`

	q = d.dbi.DB.Rebind(q)

	return client, d.dbi.DB.Select(&client, q, chain, connection_id)
}
