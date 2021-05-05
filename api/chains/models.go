package chains

import "github.com/allinbits/demeris-backend/models"

type ChainsResponse struct {
	Chains []SupportedChain `json:"chains"`
}
type SupportedChain struct {
	ChainName   string `json:"chain_name"`
	DisplayName string `json:"display_name"`
	Logo        string `json:"logo"`
}
type ChainResponse struct {
	Chain models.Chain `json:"chain"`
}

type Bech32ConfigResponse struct {
	Bech32Config models.Bech32Config `json:"bech32_config"`
}

type FeeResponse struct {
	Fee float64 `json:"fee"`
}

type FeeAddressResponse struct {
	FeeAddress string `json:"fee_address"`
}
type FeeAddress struct {
	ChainName  string `json:"chain_name"`
	FeeAddress string `json:"fee_address"`
}
type FeeAddressesResponse struct {
	FeeAddresses []FeeAddress `json:"fee_addresses"`
}

type FeeTokenResponse struct {
	FeeTokens []models.Denom `json:"fee_tokens"`
}

type PrimaryChannel struct {
	Counterparty string `json:"counterparty"`
	ChannelName  string `json:"channel_name"`
}

type PrimaryChannelResponse struct {
	Channel PrimaryChannel `json:"primary_channel"`
}
type PrimaryChannelsResponse struct {
	Channels []PrimaryChannel `json:"primary_channels"`
}

type Trace struct {
	Channel          string `json:"channel,omitempty"`
	Port             string `json:"port,omitempty"`
	ClientId         string `json:"client_id,omitempty"`
	ChainName        string `json:"chain_name,omitempty"`
	CounterpartyName string `json:"counterparty_name,omitempty"`
}

type VerifiedTrace struct {
	IbcDenom  string  `json:"ibc_denom,omitempty"`
	BaseDenom string  `json:"base_denom,omitempty"`
	Verified  bool    `json:"verified,omitempty"`
	Path      string  `json:"path,omitempty"`
	Trace     []Trace `json:"trace,omitempty"`
}
type VerifiedTraceResponse struct {
	VerifiedTrace VerifiedTrace `json:"verify_trace"`
}
