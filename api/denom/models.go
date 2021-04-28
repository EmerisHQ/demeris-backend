package denom

type Trace struct {
	Channel          string `json:"channel,omitempty"`
	Port             string `json:"port,omitempty"`
	ClientId         string `json:"client_id,omitempty"`
	ChainName        string `json:"chain_name,omitempty"`
	CounterpartyName string `json:"counterparty_name,omitempty"`
}

type VerifiedTraceResponse struct {
	IbcDenom  string  `json:"ibc_denom,omitempty"`
	BaseDenom string  `json:"base_denom,omitempty"`
	Verified  bool    `json:"verified,omitempty"`
	Path      string  `json:"path,omitempty"`
	Trace     []Trace `json:"trace,omitempty"`
}
