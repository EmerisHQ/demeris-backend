package balances

type Balances struct {
	Balances []Balance `json:"balances"`
}
type Balance struct {
	Address   string  `json:"address,omitempty"`
	BaseDenom string  `json:"base_denom,omitempty"`
	Verified  bool    `json:"verified,omitempty"`
	Amount    string  `json:"amount,omitempty"`
	OnChain   string  `json:"on_chain,omitempty"`
	Ibc       IbcInfo `json:"ibc,omitempty"`
}

type IbcInfo struct {
	Path string `json:"path,omitempty"`
	Hash string `json:"hash,omitempty"`
}
