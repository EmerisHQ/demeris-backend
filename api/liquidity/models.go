package liquidity

import sdk "github.com/cosmos/cosmos-sdk/types"

type SwapFeesResponse struct {
	Fees sdk.Coins `json:"fees"`
}
