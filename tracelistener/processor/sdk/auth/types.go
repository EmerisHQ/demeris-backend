package auth

import (
	"github.com/allinbits/demeris-backend/tracelistener"
	"github.com/cosmos/cosmos-sdk/codec"
)

type Parser interface {
	// OwnsKey returns true if key is handled by a VersionParser.
	OwnsKey(key []byte) bool

	// Process returns the address associated with data, as well as sequence number and account number for it.
	Process(cdc codec.Marshaler, data tracelistener.TraceOperation) ([]byte, uint64, uint64, error)
}
