package ibc_connections

import (
	"github.com/allinbits/demeris-backend/tracelistener"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/03-connection/types"
)

type Parser interface {
	// OwnsKey returns true if key is handled by a VersionParser.
	OwnsKey(key []byte) bool

	// Process returns a IBC ConnectionEnd and the associated ConnectionID.
	Process(cdc codec.Marshaler, data tracelistener.TraceOperation) (types.ConnectionEnd, string, error)
}
