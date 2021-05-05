package chains

import "github.com/allinbits/demeris-backend/models"

type chainsResponse struct {
	SupportedChains []string `json:"supported_chains"`
}

type chainResponse struct {
	Chain models.Chain `json:"chain"`
}

type bech32ConfigResponse struct {
	ChainName    string              `json:"chain_name"`
	Bech32Config models.Bech32Config `json:"bech32_config"`
}

type feeResponse struct {
	ChainName string  `json:"chain_name"`
	Fee       float64 `json:"fee"`
}

type feeAddressResponse struct {
	ChainName  string `json:"chain_name"`
	FeeAddress string `json:"fee_address"`
}
type feeAddressesResponse struct {
	FeeAddresses []feeAddressResponse `json:"fee_addresses"`
}

type feeTokenResponse struct {
	ChainName string         `json:"chain_name"`
	FeeTokens []models.Denom `json:"fee_tokens"`
}

type counterpartyResponse struct {
	ChainName    string `json:"chain_name"`
	Counterparty string `json:"counterparty"`
	ChannelName  string `json:"channel_name"`
}

type channelsResponse struct {
	Channels []counterpartyResponse `json:"channels"`
}

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
