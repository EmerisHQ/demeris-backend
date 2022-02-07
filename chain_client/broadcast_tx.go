package client

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/types"
	tmjson "github.com/tendermint/tendermint/libs/json"
)

// prepareBroadcast performs checks and operations before broadcasting messages
func (c *Client) PrepareBroadcast(ctx context.Context, clientCtx client.Context, msgs ...types.Msg) error {
	// validate msgs
	for _, msg := range msgs {
		if err := msg.ValidateBasic(); err != nil {
			return err
		}
	}

	c.out.Reset()

	return nil
}

// broadcast directly broadcasts the messages
func (c *Client) Broadcast(fromName string, ctx context.Context, clientCtx client.Context, msgs ...types.Msg) (*types.TxResponse, error) {
	clientCtx, err := c.BuildClientCtx(fromName)
	if err != nil {
		return &types.TxResponse{}, err
	}

	if err := c.PrepareBroadcast(ctx, clientCtx, msgs...); err != nil {
		return &types.TxResponse{}, err
	}

	// broadcast tx.
	if err := tx.BroadcastTx(clientCtx, c.factory, msgs...); err != nil {
		return &types.TxResponse{}, err
	}

	return c.handleBroadcastResult()
}

// handleBroadcastResult handles the result of broadcast messages result and checks if an error occurred
func (c *Client) handleBroadcastResult() (*types.TxResponse, error) {
	var out types.TxResponse
	if err := tmjson.Unmarshal(c.out.Bytes(), &out); err != nil {
		return &out, err
	}
	if out.Code > 0 {
		return &out, fmt.Errorf("SPN error with '%d' code: %s", out.Code, out.RawLog)
	}
	return &out, nil
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
