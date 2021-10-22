package account

import (
	"github.com/allinbits/demeris-backend/models"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
)

type balancesResponse struct {
	Balances []balance `json:"balances"`
}
type balance struct {
	Address   string  `json:"address,omitempty"`
	BaseDenom string  `json:"base_denom,omitempty"`
	Verified  bool    `json:"verified"`
	Amount    string  `json:"amount,omitempty"`
	OnChain   string  `json:"on_chain,omitempty"`
	Ibc       ibcInfo `json:"ibc,omitempty"`
}

type ibcInfo struct {
	Path string `json:"path,omitempty"`
	Hash string `json:"hash,omitempty"`
}

type stakingBalancesResponse struct {
	StakingBalances []stakingBalance `json:"staking_balances"`
}

type stakingBalance struct {
	ValidatorAddress string `json:"validator_address"`
	Amount           string `json:"amount"`
	ChainName        string `json:"chain_name"`
}

type unbondingDelegationsResponse struct {
	UnbondingDelegations []unbondingDelegation `json:"unbonding_delegations"`
}

type unbondingDelegation struct {
	ValidatorAddress string                            `json:"validator_address"`
	Entries          models.UnbondingDelegationEntries `json:"entries"`
	ChainName        string                            `json:"chain_name"`
}

type numbersResponse struct {
	Numbers []models.AuthRow `json:"numbers"`
}

type userTicketsResponse struct {
	Tickets map[string][]string `json:"tickets"`
}

type delegatorRewardsResponse struct {
	Rewards distributiontypes.QueryDelegationTotalRewardsResponse `json:"delegation_rewards"`
}
