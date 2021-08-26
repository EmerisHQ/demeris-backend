package ibc_denom_traces

import (
	"github.com/allinbits/demeris-backend/tracelistener"
	"github.com/cosmos/cosmos-sdk/codec"
	liquiditytypes "github.com/tendermint/liquidity/x/liquidity/types"
)

type Parser interface {
	// OwnsKey returns true if key is handled by a VersionParser.
	OwnsKey(key []byte) bool

	// Process returns a liquidity Pool data.
	Process(cdc codec.Marshaler, data tracelistener.TraceOperation) (liquiditytypes.Pool, error)
}
