package verifieddenom

type VerifiedDenom struct {
	Name     string `json:"name"`
	Verified bool   `json:"verified"`
}

type verifiedDenomsResponse struct {
	VerifiedDenoms []VerifiedDenom `json:"verified_denoms"`
}
