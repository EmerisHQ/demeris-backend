package v042

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"

	"github.com/allinbits/demeris-backend/tracelistener"
	"github.com/cosmos/cosmos-sdk/codec"
	host "github.com/cosmos/cosmos-sdk/x/ibc/core/24-host"
)

var ErrNotUnordered = errors.New("found channel which isn't unordered, ignoring")

type IBCChannels struct{}

func (IBCChannels) OwnsKey(key []byte) bool {
	return bytes.HasPrefix(key, []byte(host.KeyChannelEndPrefix))
}

func (IBCChannels) Process(cdc codec.Marshaler, data tracelistener.TraceOperation) (types.Channel, string, string, error) {
	var result types.Channel
	if err := cdc.UnmarshalBinaryBare(data.Value, &result); err != nil {
		return types.Channel{}, "", "", err
	}

	if err := result.ValidateBasic(); err != nil {
		return types.Channel{}, "", "", fmt.Errorf("cannot validate ibc channel, %w", err)
	}

	if result.Ordering != types.UNORDERED {
		return types.Channel{}, "", "", ErrNotUnordered
	}

	portID, channelID, err := host.ParseChannelPath(string(data.Key))
	if err != nil {
		return types.Channel{}, "", "", err
	}

	return result, portID, channelID, nil
}
