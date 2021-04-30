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
