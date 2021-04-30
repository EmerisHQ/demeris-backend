package database

import "github.com/allinbits/demeris-backend/models"

func (d *Database) QueryIBCClientTrace(chain string, channel string) (models.IbcClientInfo, error) {
	var client models.IbcClientInfo

	q := `
	SELECT 
		conn.chain_name as chain_name, 
		conn.connection_id as connection_id,
		conn.client_id as client_id,
		ch.channel_id as channel_id, 
		conn.counter_connection_id as counter_connection_id,
		conn.counter_client_id as counter_client_id,
		ch.port as port, 
		ch.state as state, 
		ch.hops as hops 
	FROM tracelistener.connections conn 
	INNER JOIN 
		(SELECT * 
			FROM tracelistener.channels 
			WHERE chain_name=? AND channel_id=?
		) ch 
	ON conn.connection_id=ANY(ch.hops);
	`

	q = d.dbi.DB.Rebind(q)

	return client, d.dbi.DB.Select(&client, q, chain, channel)
}
