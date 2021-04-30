package feeaddress

type feeAddressResponse struct {
	ChainName  string `json:"chain_name"`
	FeeAddress string `json:"fee_address"`
}
type feeAddressesResponse struct {
	FeeAddresses []feeAddressResponse `json:"fee_addresses"`
}
