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
	logo string not null,
	display_name string not null,
	primary_channel jsonb not null,
	denoms jsonb not null,
	demeris_addresses text[] not null,
	genesis_hash string not null,
	node_info jsonb not null,
	derivation_path string not null,
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
		primary_channel,
		denoms,
		demeris_addresses,
		genesis_hash,
		node_info,
		derivation_path
	)
VALUES
	(
		:chain_name,
		:enabled,
		:logo,
		:display_name,
		:valid_block_thresh,
		:primary_channel,
		:denoms,
		:demeris_addresses,
		:genesis_hash,
		:node_info,
		:derivation_path
	)
ON CONFLICT
	(chain_name)
DO UPDATE SET 
		chain_name=EXCLUDED.chain_name, 
		enabled=EXCLUDED.enabled,
		valid_block_thresh=EXCLUDED.valid_block_thresh,
		logo=EXCLUDED.logo, 
		display_name=EXCLUDED.display_name, 
		primary_channel=EXCLUDED.primary_channel, 
		denoms=EXCLUDED.denoms, 
		demeris_addresses=EXCLUDED.demeris_addresses, 
		genesis_hash=EXCLUDED.genesis_hash,
		node_info=EXCLUDED.node_info,
		derivation_path=EXCLUDED.derivation_path;
`

const getAllChains = `
SELECT * FROM cns.chains
`

const getChain = `
SELECT * FROM cns.chains WHERE chain_name='?' limit 1;
`

const channelsBetweenChains = `
select 
  c1.chain_name as c1_chain_name, 
  c1.channel_id as c1_channel_id, 
  c1.counter_channel_id as c1_counter_channel_id, 
  c2.chain_name as c2_chain_name, 
  c2.channel_id as c2_channel_id, 
  c2.counter_channel_id as c2_counter_channel_id
from 
  tracelistener.channels c1, 
  (
    select 
      chain_name, 
      channel_id, 
      counter_channel_id 
    from 
      tracelistener.channels
  ) c2 
where 
  c1.channel_id = c2.counter_channel_id 
  and c1.counter_channel_id = c2.channel_id
  and c1.chain_name != c2.chain_name
  and c1.chain_name = :source 
  and c2.chain_name = :destination;
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
