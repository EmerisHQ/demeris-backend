package ibc_denom_traces

import (
	"github.com/allinbits/demeris-backend/tracelistener"
	"github.com/cosmos/cosmos-sdk/codec"
	types2 "github.com/cosmos/cosmos-sdk/x/ibc/applications/transfer/types"
)

type Parser interface {
	// OwnsKey returns true if key is handled by a VersionParser.
	OwnsKey(key []byte) bool

	// Process returns a IBC denom trace.
	Process(cdc codec.Marshaler, data tracelistener.TraceOperation) (types2.DenomTrace, error)
}
