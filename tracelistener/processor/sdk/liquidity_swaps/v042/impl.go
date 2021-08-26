package v042

import (
	"bytes"

	liquiditytypes "github.com/tendermint/liquidity/x/liquidity/types"

	"github.com/allinbits/demeris-backend/tracelistener"
	"github.com/cosmos/cosmos-sdk/codec"
)

type LiquiditySwaps struct{}

func (LiquiditySwaps) OwnsKey(key []byte) bool {
	return bytes.HasPrefix(key, liquiditytypes.PoolKeyPrefix)
}

func (LiquiditySwaps) Process(cdc codec.Marshaler, data tracelistener.TraceOperation) (liquiditytypes.SwapMsgState, error) {
	swap := liquiditytypes.SwapMsgState{}
	if err := cdc.UnmarshalBinaryBare(data.Value, &swap); err != nil {
		return liquiditytypes.SwapMsgState{}, err
	}

	return swap, nil
}
