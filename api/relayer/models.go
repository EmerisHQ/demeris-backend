package relayer

type relayerStatusResponse struct {
	Running bool `json:"running"`
}

type relayerBalance struct {
	Address       string `json:"address"`
	ChainName     string `json:"chain_name"`
	EnoughBalance bool   `json:"enough_balance"`
}

type relayerBalances struct {
	Balances []relayerBalance `json:"balances"`
}
