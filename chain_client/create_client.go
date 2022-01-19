// Package cosmosclient provides a standalone client to connect to Cosmos SDK chains.
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"testing"

	utils "github.com/allinbits/demeris-backend/test_utils"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	staking "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/go-bip39"
	"github.com/pkg/errors"
	"github.com/tendermint/starport/starport/pkg/cosmosaccount"
	"github.com/tendermint/starport/starport/pkg/cosmosclient"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
)

const (
	defaultNodeAddress   = "http://localhost:26657"
	defaultGasAdjustment = 1.0
	defaultGasLimit      = 300000
)

// var Client cosmosaccount.Client

// Client is a client to access your chain by querying and broadcasting transactions.
type Client struct {
	// starport client
	op cosmosclient.Client

	// RPC is Tendermint RPC.
	RPC *rpchttp.HTTP `json:"rpc"`

	// Factory is a Cosmos SDK tx factory.
	Factory tx.Factory `json:"factory"`

	// context is a Cosmos SDK client context.
	Context client.Context `json:"context"`

	// AccountRegistry is the retistry to access accounts.
	AccountRegistry cosmosaccount.Registry `json:"account_registry"`

	AddressPrefix string `json:"account_address_prefix"`

	NodeAddress string `json:"node_address"`

	Mnemonic string `json:"mnemonic"`

	Key string `json:"key"`

	Denom string `json:"denom"`

	AccountAddress string `json:"account_address"`

	Out     io.Writer `json:"out"`
	ChainID string    `json:"chain_id"`

	HomePath           string                       `json:"home_path"`
	KeyringServiceName string                       `json:"keyring_service_name"`
	KeyringBackend     cosmosaccount.KeyringBackend `json:"keyring_backend"`
}

type Account cosmosaccount.Account

// Option configures your client.
type Option func(*Client)

// New creates a new client with given options.
func New(chainName string, t *testing.T, ctx context.Context, options ...Option) (Client, error) {
	var err error

	chains := utils.LoadClientChainsInfo(t)

	c := Client{
		NodeAddress:        defaultNodeAddress,
		KeyringBackend:     cosmosaccount.KeyringTest,
		AddressPrefix:      "cosmos",
		Out:                io.Discard,
		ChainID:            "test",
		KeyringServiceName: "api",
	}

	for _, ch := range chains {
		if ch.Name == chainName {

			err = json.Unmarshal(ch.Payload, &c)
			if err != nil {
				fmt.Println("Error while unmarshelling json config file : ", err)
			}

			for _, apply := range options {
				apply(&c)
			}

			if c.RPC, err = rpchttp.New(c.NodeAddress, "/websocket"); err != nil {
				return Client{}, err
			}

			statusResp, err := c.RPC.Status(ctx)
			if err != nil {
				return Client{}, err
			}

			c.ChainID = statusResp.NodeInfo.Network

			if c.HomePath == "" {
				c.HomePath = t.TempDir()
				log.Printf("Home  : %v", c.HomePath)
			}

			c.AccountRegistry, err = cosmosaccount.New(
				cosmosaccount.WithKeyringServiceName(c.KeyringServiceName),
				cosmosaccount.WithKeyringBackend(c.KeyringBackend),
				cosmosaccount.WithHome(c.HomePath),
			)
			if err != nil {
				return Client{}, err
			}

			c.Context = newContext(c.RPC, c.Out, c.ChainID, c.HomePath).WithKeyring(c.AccountRegistry.Keyring)
			c.Factory = newFactory(c.Context)

			c.AccountRegistry.Keyring, err = keyring.New(c.KeyringServiceName, string(c.KeyringBackend), c.HomePath, os.Stdin)
			if err != nil {
				return Client{}, err
			}

			return c, nil
		}
	}

	return c, nil

}

func (c Client) CreateAccount(accountName string) (acc Account, mnemonic string, err error) {
	entropySeed, err := bip39.NewEntropy(256)
	if err != nil {
		return Account{}, "", err
	}
	mnemonic, err = bip39.NewMnemonic(entropySeed)
	if err != nil {
		return Account{}, "", err
	}

	info, err := c.AccountRegistry.Keyring.NewAccount(accountName, mnemonic, "", hd.CreateHDPath(118, 0, 0).String(), hd.Secp256k1)
	if err != nil {
		return Account{}, "", err
	}

	acc = Account{
		Name: accountName,
		Info: info,
	}

	return acc, mnemonic, nil
}

func (c Client) ImportMnemonic(name, secret string) (Account, error) {
	if bip39.IsMnemonicValid(secret) {
		_, err := c.AccountRegistry.Keyring.NewAccount(name, secret, "", hd.CreateHDPath(118, 0, 0).String(), hd.Secp256k1)
		if err != nil {
			return Account{}, err
		}
	} else if err := c.AccountRegistry.Keyring.ImportPrivKey(name, secret, ""); err != nil {
		return Account{}, err
	}

	return c.GetByName(name)
}

func (c Client) GetkeysList() ([]keyring.Info, error) {
	records, err := c.AccountRegistry.Keyring.List()
	if err != nil {
		return records, err
	}

	return records, err
}

// GetByName returns an account by its name.
func (c Client) GetByName(name string) (Account, error) {
	info, err := c.AccountRegistry.Keyring.Key(name)
	if err != nil {
		return Account{}, errors.New("Key not found")
	}

	acc := Account{
		Name: name,
		Info: info,
	}

	return acc, nil
}

func (c Client) ImportKey(filePath string) error {
	bz, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	err = c.AccountRegistry.Keyring.ImportPrivKey(filePath, string(bz), "password")

	return err
}

func (c Client) Account(accountName string) (cosmosaccount.Account, error) {
	return c.AccountRegistry.GetByName(accountName)
}

func (c Client) GetBankBalances(address, denom string) (types.Coin, error) {

	var coin types.Coin

	addr, err := sdktypes.AccAddressFromBech32(address)
	if err != nil {
		return coin, err
	}

	queryClient := banktypes.NewQueryClient(c.Context)
	params := banktypes.NewQueryBalanceRequest(addr, denom)
	res, err := queryClient.Balance(context.Background(), params)
	if err != nil {
		return coin, err
	}

	out, err := c.Context.Codec.MarshalJSON(res.Balance)
	if err != nil {
		return coin, err
	}

	err = c.Context.Codec.UnmarshalJSON(out, &coin)
	if err != nil {
		return coin, err
	}

	return coin, err
}

// Address returns the account addWress from account name.
func (c Client) Address(accountName string) (sdktypes.AccAddress, error) {
	account, err := c.Account(accountName)
	if err != nil {
		return sdktypes.AccAddress{}, err
	}

	return account.Info.GetAddress(), nil
}

func newContext(
	c *rpchttp.HTTP,
	out io.Writer,
	chainID,
	home string,
) client.Context {
	var (
		amino             = codec.NewLegacyAmino()
		interfaceRegistry = codectypes.NewInterfaceRegistry()
		marshaler         = codec.NewProtoCodec(interfaceRegistry)
		txConfig          = authtx.NewTxConfig(marshaler, authtx.DefaultSignModes)
	)

	authtypes.RegisterInterfaces(interfaceRegistry)
	cryptocodec.RegisterInterfaces(interfaceRegistry)
	sdktypes.RegisterInterfaces(interfaceRegistry)
	staking.RegisterInterfaces(interfaceRegistry)
	cryptocodec.RegisterInterfaces(interfaceRegistry)

	return client.Context{}.
		WithChainID(chainID).
		WithInterfaceRegistry(interfaceRegistry).
		WithCodec(marshaler).
		WithTxConfig(txConfig).
		WithLegacyAmino(amino).
		WithInput(os.Stdin).
		WithOutput(out).
		WithAccountRetriever(authtypes.AccountRetriever{}).
		WithBroadcastMode(flags.BroadcastBlock).
		WithHomeDir(home).
		WithClient(c).
		WithSkipConfirmation(true)
}

func newFactory(clientCtx client.Context) tx.Factory {
	return tx.Factory{}.
		WithChainID(clientCtx.ChainID).
		WithKeybase(clientCtx.Keyring).
		WithGas(defaultGasLimit).
		WithGasAdjustment(defaultGasAdjustment).
		WithSignMode(signing.SignMode_SIGN_MODE_UNSPECIFIED).
		WithAccountRetriever(clientCtx.AccountRetriever).
		WithTxConfig(clientCtx.TxConfig)
}
