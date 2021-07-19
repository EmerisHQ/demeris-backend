package models

type IbcClientInfo struct {
	ChainName           string   `db:"chain_name"`
	ConnectionId        string   `db:"connection_id"`
	ClientId            string   `db:"client_id"`
	ChannelId           string   `db:"channel_id"`
	CounterConnectionID string   `db:"counter_connection_id"`
	CounterClientID     string   `db:"counter_client_id"`
	Port                string   `db:"port"`
	State               string   `db:"state"`
	Hops                []string `db:"hops"`
}

type IbcChannelInfo struct {
	ChainAName             string `db:"chain_a_chain_name"`
	ChainAChannelID        string `db:"chain_a_channel_id"`
	ChainACounterChannelID string `db:"chain_a_counter_channel_id"`
	ChainAChainID          string `db:"chain_a_chain_id"`
	ChainBName             string `db:"chain_b_chain_name"`
	ChainBChannelID        string `db:"chain_b_channel_id"`
	ChainBCounterChannelID string `db:"chain_b_counter_channel_id"`
	ChainBChainID          string `db:"chain_b_chain_id"`
}

type IbcChannelsInfo []IbcChannelInfo
