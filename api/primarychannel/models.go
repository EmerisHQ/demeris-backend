package primarychannel

type counterpartyResponse struct {
	ChainName    string `json:"chain_name"`
	Counterparty string `json:"counterparty"`
	ChannelName  string `json:"channel_name"`
}

type channelsResponse struct {
	Channels []counterpartyResponse `json:"channels"`
}
