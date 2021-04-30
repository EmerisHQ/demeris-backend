package database

import "github.com/allinbits/demeris-backend/utils/database"

const createDatabase = `
CREATE DATABASE IF NOT EXISTS cns;
`

const createTableChains = `
CREATE TABLE IF NOT EXISTS cns.chains (
	id serial unique primary key,
	chain_name string not null,
	counterparty_names jsonb not null,
	primary_channel jsonb not null,
	native_denoms jsonb not null,
	fee_tokens jsonb not null,
	fee_address text not null,
	price_modifier decimal not null,
	base_ibc_fee decimal not null,
	genesis_hash string not null
)
`

const insertChain = `
UPSERT INTO cns.chains
	(
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
