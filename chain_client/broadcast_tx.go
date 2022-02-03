package client

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/types"
)

// prepareBroadcast performs checks and operations before broadcasting messages
func (c *Client) PrepareBroadcast(ctx context.Context, clientCtx client.Context, msgs ...types.Msg) error {
	// validate msgs.
	for _, msg := range msgs {
		if err := msg.ValidateBasic(); err != nil {
			return err
		}
	}

	c.out.Reset()

	return nil
}

// broadcast directly broadcasts the messages
func (c *Client) Broadcast(ctx context.Context, clientCtx client.Context, msgs ...types.Msg) error {
	if err := c.PrepareBroadcast(ctx, clientCtx, msgs...); err != nil {
		return err
	}

	// broadcast tx.
	if err := tx.BroadcastTx(clientCtx, c.factory, msgs...); err != nil {
		return err
	}

	return c.handleBroadcastResult()
}

// handleBroadcastResult handles the result of broadcast messages result and checks if an error occurred
func (c *Client) handleBroadcastResult() error {
	out := struct {
		Code int    `json:"code"`
		Log  string `json:"raw_log"`
	}{}
	if err := json.NewDecoder(c.out).Decode(&out); err != nil {
		return err
	}
	if out.Code > 0 {
		return fmt.Errorf("SPN error with '%d' code: %s", out.Code, out.Log)
	}
	return nil
}
