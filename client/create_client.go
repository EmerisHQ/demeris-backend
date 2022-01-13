// Package cosmosclient provides a standalone client to connect to Cosmos SDK chains.
package client

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	utils "github.com/allinbits/demeris-backend/test_utils"
	"github.com/cenkalti/backoff"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	staking "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/go-bip39"
	proto "github.com/gogo/protobuf/proto"
	prototypes "github.com/gogo/protobuf/types"
	"github.com/pkg/errors"
	"github.com/tendermint/starport/starport/pkg/cosmosaccount"
	"github.com/tendermint/starport/starport/pkg/cosmosfaucet"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
)

// FaucetTransferEnsureDuration is the duration that BroadcastTx will wait when a faucet transfer
// is triggered prior to broadcasting but transfer's tx is not committed in the state yet.
var FaucetTransferEnsureDuration = time.Minute * 2

const (
	defaultNodeAddress   = "http://localhost:26657"
	defaultGasAdjustment = 1.0
	defaultGasLimit      = 300000
)

const (
	defaultFaucetAddress   = "http://localhost:4500"
	defaultFaucetDenom     = "token"
	defaultFaucetMinAmount = 100
)

// Client is a client to access your chain by querying and broadcasting transactions.
type Client struct {
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

	UseFaucet       bool   `json:"use_faucet"`
	FaucetAddress   string `json:"faucet_address"`
	FaucetDenom     string `json:"faucet_denom"`
	FaucetMinAmount uint64 `json:"faucet_min_amount"`

	HomePath           string                       `json:"home_path"`
	KeyringServiceName string                       `json:"keyring_service_name"`
	KeyringBackend     cosmosaccount.KeyringBackend `json:"keyring_backend"`
}

// Option configures your client.
type Option func(*Client)

// WithHome sets the data dir of your chain. This option is used to access your chain's
// file based keyring which is only needed when you deal with creating and signing transactions.
// when it is not provided, your data dir will be assumed as `$HOME/.your-chain-id`.
func WithHome(path string) Option {
	return func(c *Client) {
		c.HomePath = path
	}
}

// WithKeyringServiceName used as the keyring's name when you are using OS keyring backend.
// by default it is `cosmos`.
func WithKeyringServiceName(name string) Option {
	return func(c *Client) {
		c.KeyringServiceName = name
	}
}

// WithKeyringBackend sets your keyring backend. By default, it is `test`.
func WithKeyringBackend(backend cosmosaccount.KeyringBackend) Option {
	return func(c *Client) {
		c.KeyringBackend = backend
	}
}

// WithNodeAddress sets the node address of your chain. When this option is not provided
// `http://localhost:26657` is used as default.
func WithNodeAddress(addr string) Option {
	return func(c *Client) {
		c.NodeAddress = addr
	}
}

func WithAddressPrefix(prefix string) Option {
	return func(c *Client) {
		c.AddressPrefix = prefix
	}
}

func WithUseFaucet(faucetAddress, denom string, minAmount uint64) Option {
	return func(c *Client) {
		c.UseFaucet = true
		c.FaucetAddress = faucetAddress
		if denom != "" {
			c.FaucetDenom = denom
		}
		if minAmount != 0 {
			c.FaucetMinAmount = minAmount
		}
	}
}

// New creates a new client with given options.
func New(chainName string, t *testing.T, ctx context.Context, options ...Option) (Client, error) {
	var err error

	chains := utils.LoadClientChainsInfo(t)

	c := Client{
		NodeAddress:        defaultNodeAddress,
		KeyringBackend:     cosmosaccount.KeyringTest,
		AddressPrefix:      "cosmos",
		FaucetAddress:      defaultFaucetAddress,
		FaucetDenom:        defaultFaucetDenom,
		FaucetMinAmount:    defaultFaucetMinAmount,
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
				// home, err := os.UserHomeDir()
				// if err != nil {
				// 	return Client{}, err
				// }
				// c.homePath = filepath.Join(home, "."+c.chainID)
				c.HomePath = t.TempDir()
				fmt.Println("Home...", c.HomePath)
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

// Account represents an Cosmos SDK account.
type Account struct {
	// Name of the account.
	Name string

	// Info holds additional info about the account.
	Info keyring.Info
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
	fmt.Println("Key ringgggggggg.......", c.AccountRegistry.Keyring)
	records, err := c.AccountRegistry.Keyring.List()
	if err != nil {
		return records, err
	}

	return records, err
}

// GetByName returns an account by its name.
func (c Client) GetByName(name string) (Account, error) {
	info, err := c.AccountRegistry.Keyring.Key(name)
	// if errors.Is(err, dkeyring.ErrKeyNotFound) || errors.Is(err, sdkerrors.ErrKeyNotFound) {
	// 	return Account{}, &AccountDoesNotExistError{name}
	// }
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

func (c Client) GetBankBalances(address, denom string) (string, error) {

	addr, err := sdktypes.AccAddressFromBech32(address)
	if err != nil {
		return "", err
	}

	fmt.Println("context node url......", c.Context, c.Context.NodeURI)

	queryClient := banktypes.NewQueryClient(c.Context)
	params := banktypes.NewQueryBalanceRequest(addr, denom)
	res, err := queryClient.Balance(context.Background(), params)
	if err != nil {
		return "", err
	}

	// bal := c.Context.PrintProto(res.Balance)

	out, err := c.Context.Codec.MarshalJSON(res.Balance)
	if err != nil {
		return "", err
	}

	fmt.Println("BALLLLLLLLLLLLLL resp........", string(out))
	return string(out), err
}

// Address returns the account addWress from account name.
func (c Client) Address(accountName string) (sdktypes.AccAddress, error) {
	account, err := c.Account(accountName)
	if err != nil {
		return sdktypes.AccAddress{}, err
	}
	return account.Info.GetAddress(), nil
}

// Response of your broadcasted transaction.
type Response struct {
	codec codec.Codec

	// TxResponse is the underlying tx response.
	*sdktypes.TxResponse
}

// Decode decodes the proto func response defined in your Msg service into your message type.
// message needs be a pointer. and you need to provide the correct proto message(struct) type to the Decode func.
//
// e.g., for the following CreateChain func the type would be: `types.MsgCreateChainResponse`.
//
// ```proto
// service Msg {
//   rpc CreateChain(MsgCreateChain) returns (MsgCreateChainResponse);
// }
// ```
func (r Response) Decode(message proto.Message) error {
	data, err := hex.DecodeString(r.Data)
	if err != nil {
		return err
	}

	var txMsgData sdktypes.TxMsgData
	if err := r.codec.Unmarshal(data, &txMsgData); err != nil {
		return err
	}

	resData := txMsgData.Data[0]

	return prototypes.UnmarshalAny(&prototypes.Any{
		// TODO get type url dynamically(basically remove `+ "Response"`) after the following issue has solved.
		// https://github.com/cosmos/cosmos-sdk/issues/10496
		TypeUrl: resData.MsgType + "Response",
		Value:   resData.Data,
	}, message)
}

// BroadcastTx creates and broadcasts a tx with given messages for account.
func (c Client) BroadcastTx(accountName string, msgs ...sdktypes.Msg) (Response, error) {
	_, broadcast, err := c.BroadcastTxWithProvision(accountName, msgs...)
	if err != nil {
		return Response{}, err
	}
	return broadcast()
}

// protects sdktypes.Config.
var mconf sync.Mutex

func (c Client) BroadcastTxWithProvision(accountName string, msgs ...sdktypes.Msg) (
	gas uint64, broadcast func() (Response, error), err error) {
	if err := c.prepareBroadcast(context.Background(), accountName, msgs); err != nil {
		return 0, nil, err
	}

	// TODO find a better way if possible.
	mconf.Lock()
	defer mconf.Unlock()
	config := sdktypes.GetConfig()
	config.SetBech32PrefixForAccount(c.AddressPrefix, c.AddressPrefix+"pub")

	accountAddress, err := c.Address(accountName)
	if err != nil {
		return 0, nil, err
	}

	context := c.Context.
		WithFromName(accountName).
		WithFromAddress(accountAddress)

	txf, err := prepareFactory(context, c.Factory)
	if err != nil {
		return 0, nil, err
	}

	_, gas, err = tx.CalculateGas(context, txf, msgs...)
	if err != nil {
		return 0, nil, err
	}
	// the simulated gas can vary from the actual gas needed for a real transaction
	// we add an additional amount to endure sufficient gas is provided
	gas += 10000
	txf = txf.WithGas(gas)

	// Return the provision function
	return gas, func() (Response, error) {
		txUnsigned, err := tx.BuildUnsignedTx(txf, msgs...)
		if err != nil {
			return Response{}, err
		}
		if err := tx.Sign(txf, accountName, txUnsigned, true); err != nil {
			return Response{}, err
		}

		txBytes, err := context.TxConfig.TxEncoder()(txUnsigned.GetTx())
		if err != nil {
			return Response{}, err
		}

		resp, err := context.BroadcastTx(txBytes)
		return Response{
			codec:      context.Codec,
			TxResponse: resp,
		}, handleBroadcastResult(resp, err)
	}, nil
}

// prepareBroadcast performs checks and operations before broadcasting messages
func (c *Client) prepareBroadcast(ctx context.Context, accountName string, _ []sdktypes.Msg) error {
	// TODO uncomment after https://github.com/tendermint/spn/issues/363
	// validate msgs.
	//  for _, msg := range msgs {
	//  if err := msg.ValidateBasic(); err != nil {
	//  return err
	//  }
	//  }

	// account, err := c.Account(accountName)
	// if err != nil {
	// 	return err
	// }

	// // // make sure that account has enough balances before broadcasting.
	// if c.useFaucet {
	// 	// if err := c.makeSureAccountHasTokens(ctx, account.Address(c.addressPrefix)); err != nil {
	// 	// 	return err
	// 	// }
	// }

	return nil
}

// makeSureAccountHasTokens makes sure the address has a positive balance
// it requests funds from the faucet if the address has an empty balance
func (c *Client) makeSureAccountHasTokens(ctx context.Context, address string) error {
	if err := c.checkAccountBalance(ctx, address); err == nil {
		return nil
	}

	// request coins from the faucet.
	fc := cosmosfaucet.NewClient(c.FaucetAddress)
	faucetResp, err := fc.Transfer(ctx, cosmosfaucet.TransferRequest{AccountAddress: address})
	if err != nil {
		return errors.Wrap(err, "faucet server request failed")
	}
	if faucetResp.Error != "" {
		return fmt.Errorf("cannot retrieve tokens from faucet: %s", faucetResp.Error)
	}
	for _, transfer := range faucetResp.Transfers {
		if transfer.Error != "" {
			return fmt.Errorf("cannot retrieve tokens from faucet: %s", transfer.Error)
		}
	}

	// make sure funds are retrieved.
	ctx, cancel := context.WithTimeout(ctx, FaucetTransferEnsureDuration)
	defer cancel()

	return backoff.Retry(func() error {
		return c.checkAccountBalance(ctx, address)
	}, backoff.WithContext(backoff.NewConstantBackOff(time.Second), ctx))
}

func (c *Client) checkAccountBalance(ctx context.Context, address string) (err error) {
	balancesResp, err := banktypes.NewQueryClient(c.Context).AllBalances(ctx, &banktypes.QueryAllBalancesRequest{
		Address: address,
	})
	if err != nil {
		return err
	}

	// if the balance is enough do nothing.
	if len(balancesResp.Balances) > 0 {
		for _, coin := range balancesResp.Balances {
			if coin.Denom == c.FaucetDenom && coin.Amount.Uint64() >= c.FaucetMinAmount {
				return nil
			}
		}
	}

	return errors.New("account has not enough balance")
}

// handleBroadcastResult handles the result of broadcast messages result and checks if an error occurred
func handleBroadcastResult(resp *sdktypes.TxResponse, err error) error {
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return errors.New("make sure that your SPN account has enough balance")
		}

		return err
	}

	if resp.Code > 0 {
		return fmt.Errorf("SPN error with '%d' code: %s", resp.Code, resp.RawLog)
	}
	return nil
}

func prepareFactory(clientCtx client.Context, txf tx.Factory) (tx.Factory, error) {
	from := clientCtx.GetFromAddress()

	if err := txf.AccountRetriever().EnsureExists(clientCtx, from); err != nil {
		return txf, err
	}

	initNum, initSeq := txf.AccountNumber(), txf.Sequence()
	if initNum == 0 || initSeq == 0 {
		num, seq, err := txf.AccountRetriever().GetAccountNumberSequence(clientCtx, from)
		if err != nil {
			return txf, err
		}

		if initNum == 0 {
			txf = txf.WithAccountNumber(num)
		}

		if initSeq == 0 {
			txf = txf.WithSequence(seq)
		}
	}

	return txf, nil
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
