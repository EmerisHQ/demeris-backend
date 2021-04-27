package database

type Chain struct {
	Id                uint64            `db:"id"`
	ChainName         string            `db:"chain_name"`
	CounterpartyNames map[string]string `db:"counterparty_names"`
	NativeDenoms      []struct {
		IsVerified bool   `db:"is_verified"`
		Name       string `db:"name"`
	} `db:"native_denoms"`
	FeeTokens []struct {
		IsVerified bool   `db:"is_verified"`
		Name       string `db:"name"`
	} `db:"fee_tokens"`
	PriceModifier float64 `db:"price_modifier"`
	BaseIBCFee    float64 `db:"base_ibc_fee"`
	GenesisHash   string  `db:"genesis_hash"`
}

func (d *Database) Chain(name string) (Chain, error) {
	var c Chain

	n, err := d.dbi.DB.PrepareNamed("select * from cns.chains where chain_name=:name limit 1")
	if err != nil {
		return Chain{}, err
	}

	return c, n.Get(&c, map[string]interface{}{
		"name": name,
	})
}
