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
