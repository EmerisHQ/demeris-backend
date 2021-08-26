package v042

import (
	"bytes"
	"fmt"

	transferTypes "github.com/cosmos/cosmos-sdk/x/ibc/applications/transfer/types"

	"github.com/allinbits/demeris-backend/tracelistener"
	"github.com/cosmos/cosmos-sdk/codec"
)

type IBCDenomTraces struct{}

func (IBCDenomTraces) OwnsKey(key []byte) bool {
	return bytes.HasPrefix(key, transferTypes.DenomTraceKey)
}

func (IBCDenomTraces) Process(cdc codec.Marshaler, data tracelistener.TraceOperation) (transferTypes.DenomTrace, error) {
	dt := transferTypes.DenomTrace{}
	if err := cdc.UnmarshalBinaryBare(data.Value, &dt); err != nil {
		return transferTypes.DenomTrace{}, err
	}

	if err := dt.Validate(); err != nil {
		return transferTypes.DenomTrace{}, fmt.Errorf("denom trace validation failed, %w", err)
	}

	return dt, nil
}
