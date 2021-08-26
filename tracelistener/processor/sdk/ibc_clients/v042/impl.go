package v042

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
	tmIBCTypes "github.com/cosmos/cosmos-sdk/x/ibc/light-clients/07-tendermint/types"

	"github.com/allinbits/demeris-backend/tracelistener"
	"github.com/cosmos/cosmos-sdk/codec"
	host "github.com/cosmos/cosmos-sdk/x/ibc/core/24-host"
)

var ErrNotTMClientState = errors.New("found client state, but not a tendermint one, ignoring")

type IBCClients struct{}

func (IBCClients) OwnsKey(key []byte) bool {
	return bytes.Contains(key, []byte(host.KeyClientState))
}

func (IBCClients) Process(cdc codec.Marshaler, data tracelistener.TraceOperation) (*tmIBCTypes.ClientState, string, error) {
	var result exported.ClientState
	var dest *tmIBCTypes.ClientState
	if err := cdc.UnmarshalInterface(data.Value, &result); err != nil {
		return nil, "", err
	}

	if res, ok := result.(*tmIBCTypes.ClientState); !ok {
		return nil, "", ErrNotTMClientState
	} else {
		dest = res
	}

	if err := result.Validate(); err != nil {
		return nil, "", fmt.Errorf("cannot validate ibc connection, %w", err)
	}

	keySplit := strings.Split(string(data.Key), "/")
	clientID := keySplit[1]

	return dest, clientID, nil
}
