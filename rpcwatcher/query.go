package rpcwatcher

import "github.com/allinbits/demeris-backend/models"

func (w *Watcher) GetCounterParty(srcChannel string) []models.ChannelQuery {
	var c []models.ChannelQuery

	q, err := w.d.DB.PrepareNamed("select chain_name, json_data.* from cns.chains, jsonb_each_text(primary_channel) as json_data where chain_name=:chain_name and value=:channel limit 1;")
	if err != nil {
		w.l.Errorw("cannot prepare statement", "error", err)
		return []models.ChannelQuery{}
	}

	if err := q.Select(&c, map[string]interface{}{
		"chain_name": w.Name,
		"channel":    srcChannel,
	}); err != nil {
		w.l.Errorw("cannot query chain", "error", err)
	}

	return c
}