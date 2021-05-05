package delegations

type StakingBalancesResponse struct {
	StakingBalances []StakingBalance `json:"staking_balances"`
}

type StakingBalance struct {
	ValidatorAddress string `json:"validator_address"`
	Amount           string `json:"amount"`
	ChainName        string `json:"chain_name"`
}
