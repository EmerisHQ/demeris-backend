package delegations

import (
	"github.com/allinbits/demeris-backend/tracelistener"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

type Parser interface {
	// OwnsKey returns true if key is handled by a VersionParser.
	OwnsKey(key []byte) bool

	// ProcessAddition process addition of a delegation done by a user.
	// It returns the delegation object as unmarshaled from data.
	ProcessAddition(cdc codec.Marshaler, data tracelistener.TraceOperation) (types.Delegation, error)

	// ProcessRemoval processes the data for the removal of a delegation by a user.
	// It returns the delegator address and the validator address.
	ProcessRemoval(data tracelistener.TraceOperation) ([]byte, []byte, error)
}
