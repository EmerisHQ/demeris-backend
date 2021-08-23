package chains

import "github.com/allinbits/demeris-backend/models"

type chainsResponse struct {
	Chains []supportedChain `json:"chains"`
}
type supportedChain struct {
	ChainName   string `json:"chain_name"`
	DisplayName string `json:"display_name"`
	Logo        string `json:"logo"`
}
type chainResponse struct {
	Chain models.Chain `json:"chain"`
}

type bech32ConfigResponse struct {
	Bech32Config models.Bech32Config `json:"bech32_config"`
}

type feeResponse struct {
	Denoms models.DenomList `json:"denoms"`
}

type feeAddressResponse struct {
	FeeAddress []string `json:"fee_address"`
}
type feeAddress struct {
	ChainName  string   `json:"chain_name"`
	FeeAddress []string `json:"fee_address"`
}
type feeAddressesResponse struct {
	FeeAddresses []feeAddress `json:"fee_addresses"`
}

type feeTokenResponse struct {
	FeeTokens []models.Denom `json:"fee_tokens"`
}

type primaryChannel struct {
	Counterparty string `json:"counterparty"`
	ChannelName  string `json:"channel_name"`
}

type primaryChannelResponse struct {
	Channel primaryChannel `json:"primary_channel"`
}
type primaryChannelsResponse struct {
	Channels []primaryChannel `json:"primary_channels"`
}

type trace struct {
	Channel          string `json:"channel,omitempty"`
	Port             string `json:"port,omitempty"`
	ClientId         string `json:"client_id,omitempty"`
	ChainName        string `json:"chain_name,omitempty"`
	CounterpartyName string `json:"counterparty_name,omitempty"`
}

type verifiedTrace struct {
	IbcDenom  string  `json:"ibc_denom,omitempty"`
	BaseDenom string  `json:"base_denom,omitempty"`
	Verified  bool    `json:"verified,omitempty"`
	Path      string  `json:"path,omitempty"`
	Trace     []trace `json:"trace,omitempty"`
}

type verifiedTraceResponse struct {
	VerifiedTrace verifiedTrace `json:"verify_trace"`
}

type statusResponse struct {
	Online bool `json:"online"`
}

type numbersResponse struct {
	Numbers models.AuthRow `json:"numbers"`
}

type inflationResponse struct {
	Inflation string `json:"inflation"`
}

type paramsResponse struct {
	Params struct {
		MintDenom           string `json:"mint_denom"`
		InflationRateChange string `json:"inflation_rate_change"`
		InflationMax        string `json:"inflation_max"`
		InflationMin        string `json:"inflation_min"`
		GoalBonded          string `json:"goal_bonded"`
		BlocksPerYear       string `json:"blocks_per_year"`
	} `json:"params"`
}

type annualProvisionsResponse struct {
	AnnualProvisions string `json:"annual_provisions"`
}
