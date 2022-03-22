package client

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/go-bip39"
	"github.com/tendermint/starport/starport/pkg/spn"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
)

const (
	KeyringBackendTest = "test"
)

// ChainClient is client to interact with SPN.
type ChainClient struct {
	kr                 keyring.Keyring
	factory            tx.Factory
	clientCtx          client.Context
	out                *bytes.Buffer
	Address            string `json:"address"`
	AddressPrefix      string `json:"account_address_prefix"`
	RPC                string `json:"rpc"`
	Key                string `json:"key"`
	Mnemonic           string `json:"mnemonic"`
	KeyringServiceName string `json:"keyring_service_name"`
	HDPath             string `json:"hd_path"`
	Enabled            bool   `json:"enabled"`
	ChainName          string `json:"chain_name"`
	Denom              string `json:"denom"`
}

func CreateChainClient(nodeAddress, keyringServiceName, chainID, homePath string) (*ChainClient, error) {
	kr, err := keyring.New(keyringServiceName, KeyringBackendTest, homePath, os.Stdin)
	if err != nil {
		return nil, err
	}

	wsClient, err := rpchttp.New(nodeAddress, "/websocket")
	if err != nil {
		return nil, err
	}
	out := &bytes.Buffer{}
	clientCtx := spn.NewClientCtx(kr, wsClient, out, homePath).WithChainID(chainID).WithNodeURI(nodeAddress)

	factory := spn.NewFactory(clientCtx)
	return &ChainClient{
		kr:        kr,
		factory:   factory,
		clientCtx: clientCtx,
		out:       out,
	}, nil
}

// ImportMnemonic is to import existing account mnemonic in keyring
func (c ChainClient) ImportMnemonic(keyName, mnemonic, hdPath string) (acc spn.Account, err error) {
	acc, err = c.AccountCreate(keyName, mnemonic, hdPath)
	if err != nil {
		return acc, err
	}

	return acc, nil
}

// AccountCreate creates an account by name and mnemonic (optional) in the keyring.
func (c *ChainClient) AccountCreate(accountName, mnemonic, hdPath string) (spn.Account, error) {
	if mnemonic == "" {
		entropySeed, err := bip39.NewEntropy(256)
		if err != nil {
			return spn.Account{}, err
		}
		mnemonic, err = bip39.NewMnemonic(entropySeed)
		if err != nil {
			return spn.Account{}, err
		}
	}
	algos, _ := c.kr.SupportedAlgorithms()
	algo, err := keyring.NewSigningAlgoFromString(string(hd.Secp256k1Type), algos)
	if err != nil {
		return spn.Account{}, err
	}

	info, err := c.kr.NewAccount(accountName, mnemonic, "", hdPath, algo)
	if err != nil {
		return spn.Account{}, err
	}
	account := c.ToAccount(info)
	account.Mnemonic = mnemonic
	return account, nil
}

func (c ChainClient) ToAccount(info keyring.Info) spn.Account {
	ko, _ := sdktypes.Bech32ifyAddressBytes(c.AddressPrefix, info.GetAddress())

	return spn.Account{
		Name:    info.GetName(),
		Address: ko,
	}
}

// GetAccountBalances returns the balance of the account
func (c ChainClient) GetAccountBalances(address, denom string) (*types.Coin, error) {
	res, err := banktypes.NewQueryClient(c.clientCtx).
		Balance(context.Background(), &banktypes.QueryBalanceRequest{
			Address: address,
			Denom:   denom,
		})

	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, fmt.Errorf("not able to fetch balance: got response nil")
	}

	return res.Balance, err
}

// AccountList returns a list of accounts.
func (c *ChainClient) AccountList() (accounts []spn.Account, err error) {
	infos, err := c.kr.List()
	if err != nil {
		return nil, err
	}
	for _, info := range infos {
		accounts = append(accounts, c.ToAccount(info))
	}
	return accounts, nil
}

// AccountGet retrieves an account by name from the keyring.
func (c ChainClient) AccountGet(accountName string) (spn.Account, error) {
	info, err := c.kr.Key(accountName)
	if err != nil {
		return spn.Account{}, err
	}

	return c.ToAccount(info), nil
}

// GetContext return context of client
func (c *ChainClient) GetContext() client.Context {
	return c.clientCtx
}

// GetKeyring return keyring of client
func (c *ChainClient) GetKeyring() keyring.Keyring {
	return c.kr
}

// GetAccAddress return hex address from given account name
func (c *ChainClient) GetAccAddress(accountName string) (types.AccAddress, error) {
	info, err := c.clientCtx.Keyring.Key(accountName)
	if err != nil {
		return nil, err
	}

	return info.GetAddress(), nil
}

// GetBondedValidators returns bonded validators list
func (c *ChainClient) GetBondedValidators() (stakingtypes.Validators, error) {
	res, err := stakingtypes.NewQueryClient(c.clientCtx).
		Validators(context.Background(), &stakingtypes.QueryValidatorsRequest{
			Status: stakingtypes.BondStatusBonded,
		})
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, fmt.Errorf("not able to fetch validators: got response nil")
	}

	return res.Validators, err
}

// GetUnbondingDelegations returns unbonding delegations of delegator address
func (c *ChainClient) GetUnbondingDelegations(address string) (stakingtypes.UnbondingDelegations, error) {
	res, err := stakingtypes.NewQueryClient(c.clientCtx).
		DelegatorUnbondingDelegations(context.Background(), &stakingtypes.QueryDelegatorUnbondingDelegationsRequest{
			DelegatorAddr: address,
		})
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, fmt.Errorf("not able to fetch unbonding delegations: got response nil")
	}

	return res.UnbondingResponses, err
}

// GetUnbondedValidators returns bonded validators list
func (c *ChainClient) GetUnbondedValidators() (stakingtypes.Validators, error) {
	res, err := stakingtypes.NewQueryClient(c.clientCtx).
		Validators(context.Background(), &stakingtypes.QueryValidatorsRequest{
			Status: stakingtypes.BondStatusUnbonded,
		})
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, fmt.Errorf("not able to fetch validators: got response nil")
	}

	return res.Validators, err
}

// GetDelegations returns delegation balance of delegator address
func (c *ChainClient) GetDelegations(address string) (stakingtypes.DelegationResponses, error) {
	res, err := stakingtypes.NewQueryClient(c.clientCtx).
		DelegatorDelegations(context.Background(), &stakingtypes.QueryDelegatorDelegationsRequest{
			DelegatorAddr: address,
		})
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, fmt.Errorf("not able to fetch delegator delegations: got response nil")
	}

	return res.DelegationResponses, err
}
