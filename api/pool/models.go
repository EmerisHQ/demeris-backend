package pool

import sdk "github.com/cosmos/cosmos-sdk/types"

type SwapFeesResponse struct {
	Coins sdk.Coins `json:"coins"`
}
