package client

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/types"
	tmjson "github.com/tendermint/tendermint/libs/json"
)

// PrepareBroadcast performs checks and operations before broadcasting messages
func (c *ChainClient) PrepareBroadcast(msgs ...types.Msg) error {
	// validate msgs
	for _, msg := range msgs {
		if err := msg.ValidateBasic(); err != nil {
			return err
		}
	}

	c.out.Reset()

	return nil
}

// SignTx signs tx and return tx bytes
func (c *ChainClient) SignTx(fromName string, fromAddr types.AccAddress, clientCtx client.Context, msgs ...types.Msg) ([]byte, error) {
	clientCtx, err := c.BuildClientCtx(fromName, fromAddr)
	if err != nil {
		return []byte{}, err
	}

	if err := c.PrepareBroadcast(msgs...); err != nil {
		return []byte{}, err
	}

	unsignedTx, err := tx.BuildUnsignedTx(c.factory, msgs...)
	if err != nil {
		return []byte{}, err
	}

	err = tx.Sign(c.factory, clientCtx.GetFromName(), unsignedTx, true)
	if err != nil {
		return []byte{}, err
	}
	return clientCtx.TxConfig.TxEncoder()(unsignedTx.GetTx())
}

// Broadcast directly broadcasts the messages
func (c *ChainClient) Broadcast(fromName string, fromAddr types.AccAddress, clientCtx client.Context, msgs ...types.Msg) (*types.TxResponse, error) {
	clientCtx, err := c.BuildClientCtx(fromName, fromAddr)
	if err != nil {
		return &types.TxResponse{}, err
	}

	if err := c.PrepareBroadcast(msgs...); err != nil {
		return &types.TxResponse{}, err
	}

	// broadcast tx.
	if err := tx.BroadcastTx(clientCtx, c.factory, msgs...); err != nil {
		return &types.TxResponse{}, err
	}

	return c.handleBroadcastResult()
}

// handleBroadcastResult handles the result of broadcast messages result and checks if an error occurred
func (c *ChainClient) handleBroadcastResult() (*types.TxResponse, error) {
	var out types.TxResponse
	if err := tmjson.Unmarshal(c.out.Bytes(), &out); err != nil {
		return &out, err
	}
	if out.Code > 0 {
		return &out, fmt.Errorf("tx error with code '%d' log: %s", out.Code, out.RawLog)
	}
	return &out, nil
}

// BuildClientCtx builds the context for the client
func (c *ChainClient) BuildClientCtx(accountName string, accountAddress types.AccAddress) (client.Context, error) {
	return c.clientCtx.
		WithFromName(accountName).
		WithFromAddress(accountAddress), nil
}
