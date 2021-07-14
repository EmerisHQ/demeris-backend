package rpcwatcher

import "github.com/allinbits/demeris-backend/models"

func (w *Watcher) GetCounterParty(srcChannel string) ([]models.ChannelQuery, error) {
	var c []models.ChannelQuery

	q, err := w.d.DB.PrepareNamed("select chain_name, json_data.* from cns.chains, jsonb_each_text(primary_channel) as json_data where chain_name=:chain_name and value=:channel limit 1;")
	if err != nil {
		return []models.ChannelQuery{}, err
	}

	if err := q.Select(&c, map[string]interface{}{
		"chain_name": w.Name,
		"channel":    srcChannel,
	}); err != nil {
		return []models.ChannelQuery{}, err
	}

	return c, nil
}