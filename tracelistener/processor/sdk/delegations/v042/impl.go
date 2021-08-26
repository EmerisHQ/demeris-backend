package v042

import (
	"bytes"
	"fmt"

	"github.com/allinbits/demeris-backend/tracelistener"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

type Delegations struct{}

func (Delegations) OwnsKey(key []byte) bool {
	return bytes.HasPrefix(key, types.DelegationKey)
}

func (Delegations) ProcessAddition(cdc codec.Marshaler, data tracelistener.TraceOperation) (types.Delegation, error) {
	delegation := types.Delegation{}

	if err := cdc.UnmarshalBinaryBare(data.Value, &delegation); err != nil {
		return types.Delegation{}, err
	}

	return delegation, nil
}

func (Delegations) ProcessRemoval(data tracelistener.TraceOperation) ([]byte, []byte, error) {
	if len(data.Key) < 41 { // 20 bytes by address, 1 prefix = 2*20 + 1
		return nil, nil, fmt.Errorf("data mistaken for delegations data, ignoring")
	}

	delegatorAddr := data.Key[1:21]
	validatorAddr := data.Key[21:41]

	return delegatorAddr, validatorAddr, nil
}
