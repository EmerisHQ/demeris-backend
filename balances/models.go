package balances

type Balance struct {
	Address   string  `json:"address,omitempty"`
	BaseDenom string  `json:"base_denom,omitempty"`
	Verified  bool    `json:"verified,omitempty"`
	Native    bool    `json:"native,omitempty"`
	Amount    string  `json:"amount,omitempty"`
	OnChain   string  `json:"on_chain,omitempty"`
	FeeToken  bool    `json:"fee_token,omitempty"`
	Ibc       IbcInfo `json:"ibc,omitempty"`
}

type IbcInfo struct {
	SourceChain  string   `json:"source_chain,omitempty"`
	IbcDenom     string   `json:"ibc_denom,omitempty"`
	Path         string   `json:"path,omitempty"`
	VerifiedPath []string `json:"verified_path,omitempty"`
}
