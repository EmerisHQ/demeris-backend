// Package cosmosclient provides a standalone client to connect to Cosmos SDK chains.
package client

import (
	"context"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/go-bip39"
	"github.com/tendermint/starport/starport/pkg/cosmosaccount"
	"github.com/tendermint/starport/starport/pkg/cosmosclient"
)

// Client is a client to access your chain by querying and broadcasting transactions.
type Client struct {
	StarportClient     cosmosclient.Client
	AddressPrefix      string                       `json:"account_address_prefix"`
	RPC                string                       `json:"rpc"`
	Key                string                       `json:"key"`
	Mnemonic           string                       `json:"mnemonic"`
	KeyringServiceName string                       `json:"keyring_service_name"`
	KeyringBackend     cosmosaccount.KeyringBackend `json:"keyring_backend"`
	HDPath             string                       `json:"hd_path"`
}

// Option configures your client.
type Option func(*Client)

func CreateChainClient(keyringServiceName, nodeAddress, addressPrefix, homePath string) (Client, error) {
	var cli Client

	client, err := cosmosclient.New(context.Background(), cosmosclient.WithKeyringBackend(cosmosaccount.KeyringTest), cosmosclient.WithKeyringServiceName(keyringServiceName),
		cosmosclient.WithNodeAddress(nodeAddress), cosmosclient.WithAddressPrefix(addressPrefix), cosmosclient.WithHome(homePath))

	if err != nil {
		return cli, err
	}

	cli.StarportClient = client

	return cli, err
}

// CreateAccount is to create a new account
func (c Client) CreateAccount(accountName, hdPath string) (acc cosmosaccount.Account, mnemonic string, err error) {
	info, mnemonic, err := c.StarportClient.AccountRegistry.Keyring.NewMnemonic(accountName, 1, hdPath, "", hd.Secp256k1)
	if err != nil {
		return acc, "", err
	}

	acc.Name = accountName
	acc.Info = info

	return acc, mnemonic, nil
}

// ImportMnemonic is to import existing account mnemonic in keyring
func (c Client) ImportMnemonic(keyName, secret, hdPath string) (acc cosmosaccount.Account, err error) {
	if bip39.IsMnemonicValid(secret) {
		_, err := c.StarportClient.AccountRegistry.Keyring.NewAccount(keyName, secret, "", hdPath, hd.Secp256k1)
		if err != nil {
			return acc, err
		}
	} else if err := c.StarportClient.AccountRegistry.Keyring.ImportPrivKey(keyName, secret, ""); err != nil {
		return acc, err
	}

	info, err := c.StarportClient.AccountRegistry.Keyring.Key(keyName)
	if err != nil {
		return acc, err
	}

	acc.Name = keyName
	acc.Info = info

	return acc, nil
}

// GetAccountBalances returns the balance of the account
func (c Client) GetAccountBalances(address, denom string) (*types.Coin, error) {
	res, err := banktypes.NewQueryClient(c.StarportClient.Context).
		Balance(context.Background(), &banktypes.QueryBalanceRequest{
			Address: address,
			Denom:   denom,
		})

	return res.Balance, err
}
