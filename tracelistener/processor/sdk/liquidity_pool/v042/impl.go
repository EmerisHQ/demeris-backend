package v042

import (
	"bytes"

	liquiditytypes "github.com/tendermint/liquidity/x/liquidity/types"

	"github.com/allinbits/demeris-backend/tracelistener"
	"github.com/cosmos/cosmos-sdk/codec"
)

type LiquidityPool struct{}

func (LiquidityPool) OwnsKey(key []byte) bool {
	return bytes.HasPrefix(key, liquiditytypes.PoolKeyPrefix)
}

func (LiquidityPool) Process(cdc codec.Marshaler, data tracelistener.TraceOperation) (liquiditytypes.Pool, error) {
	pool := liquiditytypes.Pool{}
	if err := cdc.UnmarshalBinaryBare(data.Value, &pool); err != nil {
		return liquiditytypes.Pool{}, err
	}

	return pool, nil
}
