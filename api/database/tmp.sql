
SELECT c1.chain_name AS chain_a_chain_name,
       c1.channel_id AS chain_a_channel_id,
       c1.counter_channel_id AS chain_a_counter_channel_id,
       c1.chain_id AS chain_a_chain_id,
       c2.chain_name AS chain_b_chain_name,
       c2.channel_id AS chain_b_channel_id,
       c2.counter_channel_id AS chain_b_counter_channel_id,
       c2.chain_id AS chain_b_chain_id
FROM
    ( SELECT tracelistener.channels.chain_name,
             tracelistener.channels.channel_id,
             tracelistener.channels.counter_channel_id,
             tracelistener.clients.chain_id
     FROM tracelistener.channels
     LEFT JOIN tracelistener.connections ON tracelistener.channels.hops[1] = tracelistener.connections.connection_id
     LEFT JOIN tracelistener.clients ON tracelistener.clients.client_id = tracelistener.connections.client_id
     WHERE tracelistener.connections.chain_name = tracelistener.channels.chain_name
         AND tracelistener.clients.chain_name = tracelistener.channels.chain_name ) AS c1,

    ( SELECT tracelistener.channels.chain_name,
             tracelistener.channels.channel_id,
             tracelistener.channels.counter_channel_id,
             tracelistener.clients.chain_id
     FROM tracelistener.channels
     LEFT JOIN tracelistener.connections ON tracelistener.channels.hops[1] = tracelistener.connections.connection_id
     LEFT JOIN tracelistener.clients ON tracelistener.clients.client_id = tracelistener.connections.client_id
     WHERE tracelistener.connections.chain_name = tracelistener.channels.chain_name
         AND tracelistener.clients.chain_name = tracelistener.channels.chain_name ) AS c2
WHERE c1.channel_id = c2.counter_channel_id
    AND c1.counter_channel_id = c2.channel_id
    AND c1.chain_name != c2.chain_name
    AND c1.chain_name = 'cosmos-hub'
    AND c1.channel_id = 'channel-0'
    AND c2.chain_id = 'akash'