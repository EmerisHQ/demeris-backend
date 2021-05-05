package verifieddenoms

import "github.com/allinbits/demeris-backend/models"

type VerifiedDenomsResponse struct {
	VerifiedDenoms []VerifiedDenom `json:"verified_denoms"`
}

type VerifiedDenom struct {
	models.Denom
	ChainName string `json:"chain_name"`
}
