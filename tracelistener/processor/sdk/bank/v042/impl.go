package v042

import (
	"bytes"

	"github.com/allinbits/demeris-backend/tracelistener"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

type Bank struct{}

func (Bank) OwnsKey(key []byte) bool {
	return bytes.HasPrefix(key, types.BalancesPrefix)
}

func (Bank) Process(cdc codec.Marshaler, data tracelistener.TraceOperation) (string, sdk.Coins, error) {
	addrBytes := data.Key
	pLen := len(types.BalancesPrefix)
	addr := addrBytes[pLen : pLen+20]

	coins := sdk.Coin{
		Amount: sdk.NewInt(0),
	}

	if err := cdc.UnmarshalBinaryBare(data.Value, &coins); err != nil {
		return "", nil, err
	}

	if !coins.IsValid() {
		return "", nil, nil
	}

	return string(addr), sdk.NewCoins(coins), nil
}
