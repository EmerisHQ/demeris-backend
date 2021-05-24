package database

import "github.com/allinbits/demeris-backend/models"

func (d *Database) Chain(name string) (models.Chain, error) {
	var c models.Chain

	n, err := d.dbi.DB.PrepareNamed("select * from cns.chains where chain_name=:name and enabled=TRUE limit 1")
	if err != nil {
		return models.Chain{}, err
	}

	return c, n.Get(&c, map[string]interface{}{
		"name": name,
	})
}

func (d *Database) ChainLastBlock(name string) (models.BlockTimeRow, error) {
	var c models.BlockTimeRow

	n, err := d.dbi.DB.PrepareNamed("select * from tracelistener.blocktime where chain_name=:name and chain_name in (select chain_name from cns.chains where enabled=TRUE)")
	if err != nil {
		return models.BlockTimeRow{}, err
	}

	return c, n.Get(&c, map[string]interface{}{
		"name": name,
	})
}

func (d *Database) Chains() ([]models.Chain, error) {
	var c []models.Chain
	return c, d.dbi.Exec("select * from cns.chains where enabled=TRUE", nil, &c)
}

func (d *Database) PrimaryChannelCounterparty(chainName, counterparty string) (models.ChannelQuery, error) {
	var c models.ChannelQuery

	n, err := d.dbi.DB.PrepareNamed("select chain_name, mapping.* from cns.chains c, jsonb_each_text(primary_channel) mapping where key=:counterparty AND chain_name=:chain_name")
	if err != nil {
		return models.ChannelQuery{}, err
	}

	return c, n.Get(&c, map[string]interface{}{
		"chain_name":   chainName,
		"counterparty": counterparty,
	})
}

func (d *Database) PrimaryChannels(chainName string) ([]models.ChannelQuery, error) {
	var c []models.ChannelQuery

	n, err := d.dbi.DB.PrepareNamed("select chain_name, mapping.* from cns.chains c, jsonb_each_text(primary_channel) mapping where chain_name=:chain_name")
	if err != nil {
		return nil, err
	}

	return c, n.Select(&c, map[string]interface{}{
		"chain_name": chainName,
	})
}
