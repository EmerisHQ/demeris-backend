package ibc_clients

import (
	"github.com/allinbits/demeris-backend/tracelistener"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/ibc/light-clients/07-tendermint/types"
)

type Parser interface {
	// OwnsKey returns true if key is handled by a VersionParser.
	OwnsKey(key []byte) bool

	// Process returns the Tendermint IBC client state and client ID.
	Process(cdc codec.Marshaler, data tracelistener.TraceOperation) (*types.ClientState, string, error)
}
