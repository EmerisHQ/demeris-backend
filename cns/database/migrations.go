package database

import "github.com/allinbits/demeris-backend/utils/database"

const createDatabase = `
CREATE DATABASE IF NOT EXISTS cns;
`

const createTableChains = `
CREATE TABLE IF NOT EXISTS cns.chains (
	id serial unique primary key,
	chain_name string not null,
	logo string not null,
	display_name string not null,
	counterparty_names jsonb not null,
	primary_channel jsonb not null,
	denoms jsonb not null,
	demeris_addresses text[] not null,
	base_tx_fee jsonb not null,
	genesis_hash string not null,
	node_info jsonb not null
)
`

const deleteChain = `
DELETE FROM 
	cns.chains
WHERE
	chain_name=:chain_name; 
`

const insertChain = `
UPSERT INTO cns.chains
	(
		chain_name,
		logo,
		display_name,
		counterparty_names,
		primary_channel,
		denoms,
		demeris_addresses,
		base_tx_fee,
		genesis_hash,
		node_info
	)
VALUES
	(
		:chain_name,
		:logo,
		:display_name,
		:counterparty_names,
		:primary_channel,
		:denoms,
		:demeris_addresses,
		:base_tx_fee,
		:genesis_hash,
		:node_info
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
