package bank

import (
	"github.com/allinbits/demeris-backend/tracelistener"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Parser interface {
	// OwnsKey returns true if key is handled by a VersionParser.
	OwnsKey(key []byte) bool

	// Process returns the address associated with data, as well as the coins for the address.
	Process(cdc codec.Marshaler, data tracelistener.TraceOperation) (string, sdk.Coins, error)
}
