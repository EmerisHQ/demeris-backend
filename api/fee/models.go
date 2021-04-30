package fee

type feeResponse struct {
	ChainName string  `json:"chain_name"`
	Fee       float64 `json:"fee"`
}
