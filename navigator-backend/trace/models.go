package trace

type Trace struct {
	Path      string `json:"path,omitempty"`
	ClientId  string `json:"client_id,omitempty"`
	ChainId   bool   `json:"chain_id,omitempty"`
	ChainName bool   `json:"chain_name,omitempty"`
}

type TraceResponse struct {
	IbcDenom     string  `json:"ibc_denom,omitempty"`
	BaseDenom    string  `json:"base_denom,omitempty"`
	Verified     bool    `json:"verified,omitempty"`
	Native       bool    `json:"native,omitempty"`
	VerifiedPath string  `json:"verified_path,omitempty"`
	Trace        []Trace `json:"trace,omitempty"`
}
