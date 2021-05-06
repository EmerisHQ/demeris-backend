package verifieddenoms

import "github.com/allinbits/demeris-backend/models"

type verifiedDenom struct {
	models.Denom
	ChainName string `json:"chain_name"`
}
type verifiedDenomsResponse struct {
	VerifiedDenoms []verifiedDenom `json:"verified_denoms"`
}
