package v042

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/x/ibc/core/03-connection/types"

	"github.com/allinbits/demeris-backend/tracelistener"
	"github.com/cosmos/cosmos-sdk/codec"
	host "github.com/cosmos/cosmos-sdk/x/ibc/core/24-host"
)

var ErrNoConnection = errors.New("found connection key data but no connection, ignoring")

type IBCConnections struct{}

func (IBCConnections) OwnsKey(key []byte) bool {
	return bytes.HasPrefix(key, []byte(host.KeyConnectionPrefix))
}

func (IBCConnections) Process(cdc codec.Marshaler, data tracelistener.TraceOperation) (types.ConnectionEnd, string, error) {
	keyFields := strings.FieldsFunc(string(data.Key), func(r rune) bool {
		return r == '/'
	})

	// IBC keys are mostly strings
	switch len(keyFields) {
	case 2:
		if keyFields[0] == host.KeyConnectionPrefix { // this is a ConnectionEnd
			ce := types.ConnectionEnd{}
			if err := cdc.UnmarshalBinaryBare(data.Value, &ce); err != nil {
				return types.ConnectionEnd{}, "", fmt.Errorf("cannot unmarshal connection end, %w", err)
			}

			if err := ce.ValidateBasic(); err != nil {
				return types.ConnectionEnd{}, "", fmt.Errorf("connection end validation failed, %w", err)
			}

			return ce, keyFields[1], nil
		}
	}

	return types.ConnectionEnd{}, "", ErrNoConnection
}
