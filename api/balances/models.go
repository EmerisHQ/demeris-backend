package balances

type balancesResponse struct {
	Balances []balance `json:"balances"`
}
type balance struct {
	Address   string  `json:"address,omitempty"`
	BaseDenom string  `json:"base_denom,omitempty"`
	Verified  bool    `json:"verified,omitempty"`
	Amount    string  `json:"amount,omitempty"`
	OnChain   string  `json:"on_chain,omitempty"`
	Ibc       ibcInfo `json:"ibc,omitempty"`
}

type ibcInfo struct {
	Path string `json:"path,omitempty"`
	Hash string `json:"hash,omitempty"`
}
