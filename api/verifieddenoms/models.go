package verifieddenoms

import "github.com/allinbits/demeris-backend/models"

type verifiedDenomsResponse struct {
	VerifiedDenoms []vdEntry `json:"verified_denoms"`
}

type vdEntry struct {
	models.Denom
	ChainName string `json:"chain_name"`
}
