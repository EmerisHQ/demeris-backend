package v042

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/types/bech32"

	"github.com/allinbits/demeris-backend/tracelistener"
	"github.com/cosmos/cosmos-sdk/codec"
	types3 "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

var ErrModuleAccount = errors.New("detected module account, ignoring")

type Auth struct{}

func (Auth) OwnsKey(key []byte) bool {
	return bytes.HasPrefix(key, types.AddressStoreKeyPrefix)
}

func (Auth) Process(cdc codec.Marshaler, data tracelistener.TraceOperation) ([]byte, uint64, uint64, error) {
	if len(data.Key) != types3.AddrLen+1 {
		// key len must be len(account bytes) + 1
		return nil, 0, 0, fmt.Errorf("auth got key that isn't supposed to")
	}

	var acc types.AccountI

	if err := cdc.UnmarshalInterface(data.Value, &acc); err != nil {
		// HACK: since slashing and auth use the same prefix for two different things,
		// let's ignore "no concrete type registered for type URL *" errors.
		// This is ugly, but frankly this is the only way to do it.
		// Frojdi please bless us with the new SDK ASAP.

		if strings.HasPrefix(err.Error(), "no concrete type registered for type URL") {
			return nil, 0, 0, fmt.Errorf("value for account isn't accountI")
		}

		return nil, 0, 0, err
	}

	if _, ok := acc.(*types.ModuleAccount); ok {
		// ignore moduleaccounts
		return nil, 0, 0, ErrModuleAccount
	}

	baseAcc, ok := acc.(*types.BaseAccount)
	if !ok {
		return nil, 0, 0, fmt.Errorf("cannot cast account to BaseAccount, type %T, account object type %T", baseAcc, acc)
	}

	if err := baseAcc.Validate(); err != nil {
		return nil, 0, 0, fmt.Errorf("non compliant auth account, %w", err)
	}

	_, bz, err := bech32.DecodeAndConvert(baseAcc.Address)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("cannot parse %s as bech32, %w", baseAcc.Address, err)
	}

	return bz, acc.GetAccountNumber(), acc.GetSequence(), nil

}
