package client

import (
	"bytes"
	"context"
	"os"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/go-bip39"
	"github.com/tendermint/starport/starport/pkg/spn"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
)

// Client is client to interact with SPN.
type Client struct {
	SpnClient          spn.Client
	kr                 keyring.Keyring
	factory            tx.Factory
	clientCtx          client.Context
	out                *bytes.Buffer
	AddressPrefix      string `json:"account_address_prefix"`
	RPC                string `json:"rpc"`
	Key                string `json:"key"`
	Mnemonic           string `json:"mnemonic"`
	KeyringServiceName string `json:"keyring_service_name"`
	HDPath             string `json:"hd_path"`
	Enabled            bool   `json:"enabled"`
	ChainName          string `json:"chain_name"`
}

func CreateChainClient(nodeAddress, keyrimgServiceName, homePath string) (*Client, error) {
	kr, err := keyring.New(keyrimgServiceName, "test", homePath, os.Stdin)
	if err != nil {
		return nil, err
	}

	client, err := rpchttp.New(nodeAddress, "/websocket")
	if err != nil {
		return nil, err
	}
	out := &bytes.Buffer{}
	clientCtx := spn.NewClientCtx(kr, client, out, homePath)
	factory := spn.NewFactory(clientCtx)
	return &Client{
		kr:        kr,
		factory:   factory,
		clientCtx: clientCtx,
		out:       out,
	}, nil
}

// ImportMnemonic is to import existing account mnemonic in keyring
func (c Client) ImportMnemonic(keyName, mnemonic, hdPath string) (acc spn.Account, err error) {
	acc, err = c.AccountCreate(keyName, mnemonic, hdPath)
	if err != nil {
		return acc, err
	}

	return acc, nil
}

// AccountCreate creates an account by name and mnemonic (optional) in the keyring.
func (c *Client) AccountCreate(accountName, mnemonic, hdPath string) (spn.Account, error) {
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
	account := toAccount(info)
	account.Mnemonic = mnemonic
	return account, nil
}

func toAccount(info keyring.Info) spn.Account {
	ko, _ := keyring.Bech32KeyOutput(info)
	return spn.Account{
		Name:    ko.Name,
		Address: ko.Address,
	}
}

// GetAccountBalances returns the balance of the account
func (c Client) GetAccountBalances(address, denom string) (*types.Coin, error) {
	res, err := banktypes.NewQueryClient(c.clientCtx).
		Balance(context.Background(), &banktypes.QueryBalanceRequest{
			Address: address,
			Denom:   denom,
		})

	return res.Balance, err
}

// AccountList returns a list of accounts.
func (c *Client) AccountList() (accounts []spn.Account, err error) {
	infos, err := c.kr.List()
	if err != nil {
		return nil, err
	}
	for _, info := range infos {
		accounts = append(accounts, toAccount(info))
	}
	return accounts, nil
}

// AccountGet retrieves an account by name from the keyring.
func (c *Client) AccountGet(accountName string) (spn.Account, error) {
	info, err := c.kr.Key(accountName)
	if err != nil {
		return spn.Account{}, err
	}
	return toAccount(info), nil
}

// buildClientCtx builds the context for the client
func (c *Client) BuildClientCtx(accountName string) (client.Context, error) {
	info, err := c.kr.Key(accountName)
	if err != nil {
		return client.Context{}, err
	}
	return c.clientCtx.
		WithFromName(accountName).
		WithFromAddress(info.GetAddress()), nil
}
