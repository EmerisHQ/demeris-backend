package database

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
SELECT
c1.chain_name AS chain_a_chain_name,
c1.channel_id AS chain_a_channel_id,
c1.counter_channel_id AS chain_a_counter_channel_id,
c1.chain_id AS chain_a_chain_id,
c2.chain_name AS chain_b_chain_name,
c2.channel_id AS chain_b_channel_id,
c2.counter_channel_id AS chain_b_counter_channel_id,
c2.chain_id AS chain_b_chain_id
FROM
(
SELECT
tracelistener.channels.chain_name,
tracelistener.channels.channel_id,
tracelistener.channels.counter_channel_id,
tracelistener.clients.chain_id
FROM
tracelistener.channels
LEFT JOIN tracelistener.connections ON
tracelistener.channels.hops[1]
= tracelistener.connections.connection_id
LEFT JOIN tracelistener.clients ON
tracelistener.clients.client_id
= tracelistener.connections.client_id
WHERE
tracelistener.connections.chain_name
= tracelistener.channels.chain_name
AND tracelistener.clients.chain_name
= tracelistener.channels.chain_name
)
AS c1,
(
SELECT
tracelistener.channels.chain_name,
tracelistener.channels.channel_id,
tracelistener.channels.counter_channel_id,
tracelistener.clients.chain_id
FROM
tracelistener.channels
LEFT JOIN tracelistener.connections ON
tracelistener.channels.hops[1]
= tracelistener.connections.connection_id
LEFT JOIN tracelistener.clients ON
tracelistener.clients.client_id
= tracelistener.connections.client_id
WHERE
tracelistener.connections.chain_name
= tracelistener.channels.chain_name
AND tracelistener.clients.chain_name
= tracelistener.channels.chain_name
)
AS c2
WHERE
c1.channel_id = c2.counter_channel_id
AND c1.counter_channel_id = c2.channel_id
AND c1.chain_name != c2.chain_name
AND c1.chain_name = :source
AND c1.channel_id = :destination
AND c2.chain_id = :chainID
`
