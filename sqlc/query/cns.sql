-- name: Chain :one
SELECT * FROM chains WHERE chain_name=$1 limit 1;

-- name: AddChain :exec
INSERT INTO chains
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
    derivation_path,
    block_explorer
)
VALUES
(
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9,
    $10,
    $11,
    $12
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
        derivation_path=EXCLUDED.derivation_path,
        block_explorer=EXCLUDED.block_explorer
    RETURNING chain_name;

-- name: ChainAmount :one
select count(id) from chains;

-- name: Chains :many
SELECT * FROM chains;

-- name: UpdateDenoms :exec
UPDATE chains
	SET denoms = $1
	WHERE chain_name=$2;


/* -- name: ChannelsBetweenChains :many
SELECT
    c1.chain_name AS chain_a_chain_name,
    c1.channel_id AS chain_a_channel_id,
    c1.counter_channel_id AS chain_a_counter_channel_id,
    c1.chain_id AS chain_a_chain_id,
    c1.state AS chain_a_state,
    c2.chain_name AS chain_b_chain_name,
    c2.channel_id AS chain_b_channel_id,
    c2.counter_channel_id AS chain_b_counter_channel_id,
    c2.chain_id AS chain_b_chain_id,
    c2.state AS chain_b_state
FROM
    (
        SELECT
            tracelistener.channels.chain_name,
            tracelistener.channels.channel_id,
            tracelistener.channels.counter_channel_id,
            tracelistener.clients.chain_id,
            tracelistener.channels.state
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
            tracelistener.clients.chain_id,
            tracelistener.channels.state
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
  AND c1.state = '3'
  AND c2.state = '3'
  AND c1.chain_name = $1
  AND c2.chain_name = $2
  AND c2.chain_id = $3; */