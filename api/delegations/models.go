package delegations

type stakingBalancesResponse struct {
	StakingBalances []stakingBalance `json:"staking_balances"`
}

type stakingBalance struct {
	ValidatorAddress string `json:"validator_address"`
	Amount           string `json:"amount"`
	ChainName        string `json:"chain_name"`
}
