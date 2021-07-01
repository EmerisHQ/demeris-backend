package database

import (
	"fmt"

	"github.com/allinbits/demeris-backend/models"
)

func (d *Database) GetIbcChannelToChain(chain string, channel string) (models.IbcChannelsInfo, error) {
	var c models.IbcChannelsInfo

	q := `
	SELECT 
		c1.chain_name as chain_a_chain_name, 
		c1.channel_id as chain_a_channel_id, 
		c1.counter_channel_id as chain_a_counter_channel_id, 
		c2.chain_name as chain_b_chain_name, 
		c2.channel_id as chain_b_channel_id, 
		c2.counter_channel_id as chain_b_counter_channel_id
	FROM 
		tracelistener.channels c1, 
		(
		SELECT 
			chain_name, 
			channel_id, 
			counter_channel_id 
		FROM 
			tracelistener.channels
		) c2 
	WHERE 
		c1.channel_id = c2.counter_channel_id 
		AND c1.counter_channel_id = c2.channel_id 
	    AND c1.chain_name != c2.chain_name
		AND c1.chain_name = ?
		AND c1.channel_id = ?;
	`

	q = d.dbi.DB.Rebind(q)

	err := d.dbi.DB.Select(&c, q, chain, channel)
	if err != nil {
		return nil, err
	}

	if len(c) == 0 {
		return nil, fmt.Errorf("query done but returned no result")
	}

	return c, nil
}
