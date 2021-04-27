package feetoken

type FeeToken struct {
	Name     string `json:"name"`
	Verified bool   `json:"verified"`
}

type feeTokenResponse struct {
	FeeTokens []FeeToken `json:"fee_tokens"`
}
