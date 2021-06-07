package database

import "github.com/allinbits/demeris-backend/utils/database"

const createDatabase = `
CREATE DATABASE IF NOT EXISTS cns;
`

const createTableChains = `
CREATE TABLE IF NOT EXISTS cns.chains (
	id serial unique primary key,
	enabled boolean default false,
	chain_name string not null,
	valid_block_thresh string not null,
	minimum_thresh_relayer_balance bigint not null,
	logo string not null,
	display_name string not null,
	counterparty_names jsonb not null,
	primary_channel jsonb not null,
	denoms jsonb not null,
	demeris_addresses text[] not null,
	base_tx_fee jsonb not null,
	genesis_hash string not null,
	node_info jsonb not null,
	unique(chain_name)
)
`

const deleteChain = `
DELETE FROM 
	cns.chains
WHERE
	chain_name=:chain_name; 
`

const insertChain = `
INSERT INTO cns.chains
	(
		chain_name,
		enabled,
		logo,
		display_name,
		valid_block_thresh,
		minimum_thresh_relayer_balance,
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
		:enabled,
		:logo,
		:display_name,
		:valid_block_thresh,
		:minimum_thresh_relayer_balance,
		:counterparty_names,
		:primary_channel,
		:denoms,
		:demeris_addresses,
		:base_tx_fee,
		:genesis_hash,
		:node_info
	)
ON CONFLICT
	(chain_name)
DO UPDATE SET 
		chain_name=EXCLUDED.chain_name, 
		enabled=EXCLUDED.enabled,
		valid_block_thresh=EXCLUDED.valid_block_thresh,
		minimum_thresh_relayer_balance=EXCLUDED.minimum_thresh_relayer_balance,
		logo=EXCLUDED.logo, 
		display_name=EXCLUDED.display_name, 
		counterparty_names=EXCLUDED.counterparty_names, 
		primary_channel=EXCLUDED.primary_channel, 
		denoms=EXCLUDED.denoms, 
		demeris_addresses=EXCLUDED.demeris_addresses, 
		base_tx_fee=EXCLUDED.base_tx_fee,
		genesis_hash=EXCLUDED.genesis_hash,
		node_info=EXCLUDED.node_info;
`

const getAllChains = `
SELECT * FROM cns.chains
`

const getChain = `
SELECT * FROM cns.chains WHERE chain_name='?' limit 1;
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
