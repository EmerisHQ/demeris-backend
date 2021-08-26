package ibc_channels

import (
	"github.com/allinbits/demeris-backend/tracelistener"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
)

type Parser interface {
	// OwnsKey returns true if key is handled by a VersionParser.
	OwnsKey(key []byte) bool

	// Process returns the IBC channel, port ID and channel ID associated with data.
	Process(cdc codec.Marshaler, data tracelistener.TraceOperation) (types.Channel, string, string, error)
}
