package block

import (
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
)

type blockHeightResp struct {
	JSONRPC string                `json:"jsonrpc"`
	ID      string                `json:"id"`
	Result  coretypes.ResultBlock `json:"result"`
}
