// Package cosmosclient provides a standalone client to connect to Cosmos SDK chains.
package client

import (
	"context"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/go-bip39"
	"github.com/pkg/errors"
	"github.com/tendermint/starport/starport/pkg/cosmosaccount"
	"github.com/tendermint/starport/starport/pkg/cosmosclient"
)

// Client is a client to access your chain by querying and broadcasting transactions.
type Client struct {
	starportClient     cosmosclient.Client
	AddressPrefix      string                       `json:"account_address_prefix"`
	NodeAddress        string                       `json:"node_address"`
	Mnemonic           string                       `json:"mnemonic"`
	KeyringServiceName string                       `json:"keyring_service_name"`
	KeyringBackend     cosmosaccount.KeyringBackend `json:"keyring_backend"`
	HDPath             string                       `json:"hd_path"`
}

type Account cosmosaccount.Account

// Option configures your client.
type Option func(*Client)

func CreateChainClient(keyringServiceName, nodeAddress, addressPrefix, homePath string) (Client, error) {
	var cli Client

	client, err := cosmosclient.New(context.Background(), cosmosclient.WithKeyringBackend(cosmosaccount.KeyringTest), cosmosclient.WithKeyringServiceName(keyringServiceName),
		cosmosclient.WithNodeAddress(nodeAddress), cosmosclient.WithAddressPrefix(addressPrefix), cosmosclient.WithHome(homePath))

	if err != nil {
		return cli, err
	}

	cli.starportClient = client

	return cli, err
}

// CreateAccount is to create a new account
func (c Client) CreateAccount(accountName, hdPath string) (acc Account, mnemonic string, err error) {
	entropySeed, err := bip39.NewEntropy(256)
	if err != nil {
		return Account{}, "", err
	}
	mnemonic, err = bip39.NewMnemonic(entropySeed)
	if err != nil {
		return Account{}, "", err
	}

	info, err := c.starportClient.AccountRegistry.Keyring.NewAccount(accountName, mnemonic, "", hdPath, hd.Secp256k1)
	if err != nil {
		return Account{}, "", err
	}

	acc = Account{
		Name: accountName,
		Info: info,
	}

	return acc, mnemonic, nil
}

// ImportMnemonic is to import existing account mnemonic in keyring
func (c Client) ImportMnemonic(name, secret, hdPath string) (Account, error) {
	if bip39.IsMnemonicValid(secret) {
		_, err := c.starportClient.AccountRegistry.Keyring.NewAccount(name, secret, "", hdPath, hd.Secp256k1)
		if err != nil {
			return Account{}, err
		}
	} else if err := c.starportClient.AccountRegistry.Keyring.ImportPrivKey(name, secret, ""); err != nil {
		return Account{}, err
	}

	return c.GetByName(name)
}

// GetkeysList returns the list of keys
func (c Client) GetkeysList() ([]keyring.Info, error) {
	records, err := c.starportClient.AccountRegistry.Keyring.List()
	if err != nil {
		return records, err
	}

	return records, err
}

// GetByName returns an account by its name.
func (c Client) GetByName(name string) (Account, error) {
	info, err := c.starportClient.AccountRegistry.Keyring.Key(name)
	if err != nil {
		return Account{}, errors.New("Key not found")
	}

	acc := Account{
		Name: name,
		Info: info,
	}

	return acc, nil
}

// GetAccountBalances returns the balance of the account
func (c Client) GetAccountBalances(address, denom string) (types.Coin, error) {
	var coin types.Coin

	addr, err := sdktypes.AccAddressFromBech32(address)
	if err != nil {
		return coin, err
	}

	queryClient := banktypes.NewQueryClient(c.starportClient.Context)
	params := banktypes.NewQueryBalanceRequest(addr, denom)
	res, err := queryClient.Balance(context.Background(), params)
	if err != nil {
		return coin, err
	}

	out, err := c.starportClient.Context.Codec.MarshalJSON(res.Balance)
	if err != nil {
		return coin, err
	}

	err = c.starportClient.Context.Codec.UnmarshalJSON(out, &coin)
	if err != nil {
		return coin, err
	}

	return coin, err
}
