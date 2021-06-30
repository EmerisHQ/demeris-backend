package relayer

type relayerStatusResponse struct {
	Running bool `json:"running"`
}

type relayerBalance struct {
	Address       string `json:"address"`
	EnoughBalance bool   `json:"enough_balance"`
}

type relayerBalances []relayerBalance
