package tests

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	chainClient "github.com/allinbits/demeris-backend/chain_client"
	"github.com/allinbits/demeris-backend/test_utils"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctransfertypes "github.com/cosmos/cosmos-sdk/x/ibc/applications/transfer/types"
	ibcclienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
)

const (
	verifyTraceEndpoint = "/chain/%s/denom/verify_trace/%s"
)

func (suite *testCtx) TestVerifyTrace() {
	var enabledChains []test_utils.EnvChain
	for _, chain := range suite.Chains {
		if chain.Enabled {
			enabledChains = append(enabledChains, chain)
		}
	}
	var chainA, chainB test_utils.EnvChain
	for {
		a := rand.Intn(len(enabledChains))
		b := rand.Intn(len(enabledChains))
		if a != b {
			chainA = enabledChains[a]
			chainB = enabledChains[b]
			break
		}
	}

	var ccA, ccB chainClient.Client
	for _, ch := range suite.clientChains {
		if ch.Name == chainA.Name {
			err := json.Unmarshal(ch.Payload, &ccA)
			suite.Require().NoError(err)
		} else if ch.Name == chainB.Name {
			err := json.Unmarshal(ch.Payload, &ccB)
			suite.Require().NoError(err)
		}
	}
	cliB := chainClient.GetClient(suite.T(), suite.Env, chainB.Name, ccB)
	suite.Require().NotNil(cliB)
	rec_account, err := cliB.AccountGet(ccB.Key)
	suite.Require().NoError(err)

	cliA := chainClient.GetClient(suite.T(), suite.Env, chainA.Name, ccA)
	suite.Require().NotNil(cliA)
	send_account, err := cliA.AccountGet(ccA.Key)
	suite.Require().NoError(err)
	fromAddr, err := sdk.AccAddressFromBech32(send_account.Address)
	suite.Require().NoError(err)

	var chainAData map[string]interface{}
	suite.Require().NoError(json.Unmarshal(chainA.Payload, &chainAData))

	channelBytes, err := json.Marshal(chainAData["primary_channel"])
	suite.Require().NoError(err)
	var primary_channels map[string]string
	suite.Require().NoError(json.Unmarshal(channelBytes, &primary_channels))

	denomBytes, err := json.Marshal(chainAData["denoms"])
	suite.Require().NoError(err)
	var denoms []map[string]interface{}
	suite.Require().NoError(json.Unmarshal(denomBytes, &denoms))

	token := sdk.Coin{
		Denom:  denoms[0]["name"].(string),
		Amount: sdk.NewInt(100),
	}
	resp, err := http.Get(ccB.RPC + "/block?height")
	suite.Require().NoError(err)

	bz, err := ioutil.ReadAll(resp.Body)
	suite.Require().NoError(err)
	var blockData map[string]interface{}
	suite.Require().NoError(json.Unmarshal(bz, &blockData))
	heightFromResp := blockData["result"].(map[string]interface{})["block"].(map[string]interface{})["header"].(map[string]interface{})["height"].(string)
	height, err := strconv.Atoi(heightFromResp)
	suite.Require().NoError(err)
	timeoutHeight := ibcclienttypes.Height{
		RevisionNumber: 0,
		RevisionHeight: uint64(height + 100),
	}
	ibcDenom := strings.ToUpper(fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("transfer/%s/%s", primary_channels[ccB.ChainName], denoms[0]["name"].(string))))))
	prevBalance, err := cliB.GetAccountBalances(rec_account.Address, fmt.Sprintf("ibc/%s", ibcDenom))
	suite.Require().NoError(err)

	msg := ibctransfertypes.NewMsgTransfer("transfer", primary_channels[ccB.ChainName], token, fromAddr, rec_account.Address, timeoutHeight, 0)

	_, err = cliA.Broadcast(ccA.Key, context.Background(), cliA.GetContext(), msg)
	suite.Require().NoError(err)

	time.Sleep(time.Second * 8)

	postBalance, err := cliB.GetAccountBalances(rec_account.Address, fmt.Sprintf("ibc/%s", ibcDenom))
	suite.Require().NoError(err)

	suite.Require().Equal(int64(100), postBalance.Amount.Int64()-prevBalance.Amount.Int64())

}
