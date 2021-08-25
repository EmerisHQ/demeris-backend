package database

import "github.com/allinbits/demeris-backend/models"

func (d *Database) Chain(name string) (models.Chain, error) {
	var c models.Chain

	n, err := d.dbi.DB.PrepareNamed("select * from cns.chains where chain_name=:name and enabled=TRUE limit 1")
	if err != nil {
		return models.Chain{}, err
	}

	defer func() {
		err := n.Close()
		if err != nil {
			panic(err)
		}
	}()

	return c, n.Get(&c, map[string]interface{}{
		"name": name,
	})
}

func (d *Database) ChainFromChainID(chainID string) (models.Chain, error) {
	var c models.Chain

	n, err := d.dbi.DB.PrepareNamed("select * from cns.chains where node_info->>'chain_id'=:chainID and enabled=TRUE limit 1;")
	if err != nil {
		return models.Chain{}, err
	}

	defer func() {
		err := n.Close()
		if err != nil {
			panic(err)
		}
	}()

	return c, n.Get(&c, map[string]interface{}{
		"chainID": chainID,
	})
}

func (d *Database) ChainLastBlock(name string) (models.BlockTimeRow, error) {
	var c models.BlockTimeRow

	n, err := d.dbi.DB.PrepareNamed("select * from tracelistener.blocktime where chain_name=:name and chain_name in (select chain_name from cns.chains where enabled=TRUE)")
	if err != nil {
		return models.BlockTimeRow{}, err
	}

	defer func() {
		err := n.Close()
		if err != nil {
			panic(err)
		}
	}()

	return c, n.Get(&c, map[string]interface{}{
		"name": name,
	})
}

func (d *Database) Chains() ([]models.Chain, error) {
	var c []models.Chain
	return c, d.dbi.Exec("select * from cns.chains where enabled=TRUE", nil, &c)
}

func (d *Database) VerifiedDenoms() (map[string]models.DenomList, error) {
	var c []models.Chain
	if err := d.dbi.Exec("select chain_name, denoms from cns.chains where enabled=TRUE", nil, &c); err != nil {
		return nil, err
	}

	ret := make(map[string]models.DenomList)

	for _, cc := range c {
		ret[cc.ChainName] = cc.VerifiedTokens()
	}

	return ret, nil
}

func (d *Database) SimpleChains() ([]models.Chain, error) {
	var c []models.Chain
	return c, d.dbi.Exec("select chain_name, display_name, logo from cns.chains where enabled=TRUE", nil, &c)
}

func (d *Database) ChainIDs() (map[string]string, error) {
	type it struct {
		ChainName string `db:"chain_name"`
		ChainID   string `db:"chain_id"`
	}

	c := map[string]string{}
	var cc []it
	err := d.dbi.Exec("select chain_name, node_info->>'chain_id' as chain_id from cns.chains where enabled=TRUE", nil, &cc)
	if err != nil {
		return nil, err
	}

	for _, ccc := range cc {
		c[ccc.ChainName] = ccc.ChainID
	}

	return c, nil
}

func (d *Database) PrimaryChannelCounterparty(chainName, counterparty string) (models.ChannelQuery, error) {
	var c models.ChannelQuery

	n, err := d.dbi.DB.PrepareNamed("select chain_name, mapping.* from cns.chains c, jsonb_each_text(primary_channel) mapping where key=:counterparty AND chain_name=:chain_name")
	if err != nil {
		return models.ChannelQuery{}, err
	}

	defer func() {
		err := n.Close()
		if err != nil {
			panic(err)
		}
	}()

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

	defer func() {
		err := n.Close()
		if err != nil {
			panic(err)
		}
	}()

	return c, n.Select(&c, map[string]interface{}{
		"chain_name": chainName,
	})
}
