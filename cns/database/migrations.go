package database

import "github.com/allinbits/demeris-backend/utils/database"

const createDatabase = `
CREATE DATABASE IF NOT EXISTS cns;
`

const createTableChains = `
CREATE TABLE IF NOT EXISTS cns.chains (
	id serial not null,
	chain_name string not null,
	counterparty_names jsonb not null,
	native_denoms jsonb not null,
	fee_tokens jsonb not null,
	price_modifier decimal not null,
	base_ibc_fee decimal not null,
	genesis_hash string not null,
	unique(id)
)
`

const insertChain = `
UPSERT INTO cns.chains
	(
		id,
		chain_name,
		counterparty_names,
		native_denoms,
		fee_tokens,
		price_modifier,
		base_ibc_fee,
		genesis_hash
	)
VALUES
	(
		:id,
		:chain_name,
		:counterparty_names,
		:native_denoms,
		:fee_tokens,
		:price_modifier,
		:base_ibc_fee,
		:genesis_hash
	)
`

const getAllChains = `
SELECT * FROM cns.chains
`

var migrationList = []string{
	createDatabase,
	createTableChains,
}

func (i *Instance) runMigrations() {
	if err := database.RunMigrations(i.connString, migrationList); err != nil {
		panic(err)
	}
}
