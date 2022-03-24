package tests

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	suite.T().Parallel()
	// filter enabled chains
	var enabledChains []test_utils.EnvChain
	for _, chain := range suite.Chains {
		if chain.Enabled {
			enabledChains = append(enabledChains, chain)
		}
	}
	suite.Require().Greater(len(enabledChains), 1, "Need atleast 2 enabled chains to perform IBC transaction")

	// pick 2 random chains
	var chainA, chainB test_utils.EnvChain
	for {
		// a := rand.Intn(len(enabledChains))
		// b := rand.Intn(len(enabledChains))
		a := 0
		b := 1
		if a != b {
			chainA = enabledChains[a]
			chainB = enabledChains[b]
			break
		}
	}

	// create clients and accounts for above picked chains
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

	// get respective channel for chainB from chainA payload
	var chainAData map[string]interface{}
	suite.Require().NoError(json.Unmarshal(chainA.Payload, &chainAData))

	primary_channels := chainAData["primary_channel"].(map[string]interface{})

	// get chainA denom
	denomBytes, err := json.Marshal(chainAData["denoms"])
	suite.Require().NoError(err)
	var denoms []map[string]interface{}
	suite.Require().NoError(json.Unmarshal(denomBytes, &denoms))

	// check balance for account A
	accABalance, err := cliA.GetAccountBalances(send_account.Address, denoms[0]["name"].(string))
	suite.Require().NoError(err)
	suite.Require().Greater(accABalance.Amount.BigInt().Uint64(), uint64(100), "Not enough balance to make an IBC transaction")

	token := sdk.Coin{
		Denom:  denoms[0]["name"].(string),
		Amount: sdk.NewInt(100),
	}

	// get current block height and set timeout height
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

	// build IBC denom hash
	ibcDenom := strings.ToUpper(fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("transfer/%s/%s", primary_channels[ccB.ChainName].(string), denoms[0]["name"].(string))))))

	// get account B balance before IBC transaction
	prevBalance, err := cliB.GetAccountBalances(rec_account.Address, fmt.Sprintf("ibc/%s", ibcDenom))
	suite.Require().NoError(err)

	// build and broadcast ibc transfer message
	msg := ibctransfertypes.NewMsgTransfer("transfer", primary_channels[ccB.ChainName].(string), token, fromAddr, rec_account.Address, timeoutHeight, 0)

	_, err = cliA.Broadcast(ccA.Key, context.Background(), cliA.GetContext(), msg)
	suite.Require().NoError(err)

	time.Sleep(time.Second * 10)

	// get account B balance after IBC transaction
	postBalance, err := cliB.GetAccountBalances(rec_account.Address, fmt.Sprintf("ibc/%s", ibcDenom))
	suite.Require().NoError(err)

	// check updated balance
	suite.Require().Equal(uint64(100), postBalance.Amount.BigInt().Uint64()-prevBalance.Amount.BigInt().Uint64())

	url := suite.Client.BuildUrl(verifyTraceEndpoint, chainB.Name, ibcDenom)
	// act
	respTrace, err := suite.Client.Get(url)
	suite.Require().NoError(err)
	suite.Require().Equal(http.StatusOK, respTrace.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", chainB.Name, resp.StatusCode))

	body, err := ioutil.ReadAll(respTrace.Body)
	suite.Require().NoError(err)
	var result map[string]interface{}
	suite.Require().NoError(json.Unmarshal(body, &result))

	var chainBData map[string]interface{}
	suite.Require().NoError(json.Unmarshal(chainB.Payload, &chainBData))
	primary_channels = chainBData["primary_channel"].(map[string]interface{})
	fmt.Println(result)
	suite.Require().Equal(result["verify_trace"].(map[string]interface{})["base_denom"].(string), denoms[0]["name"].(string))
	suite.Require().Equal(result["verify_trace"].(map[string]interface{})["path"].(string), fmt.Sprintf("transfer/%s", primary_channels[chainA.Name].(string)))
	suite.Require().Equal(result["verify_trace"].(map[string]interface{})["verified"].(bool), denoms[0]["verified"].(bool))
	suite.Require().Equal(result["verify_trace"].(map[string]interface{})["trace"].(map[string]interface{})["chain_name"].(string), chainB.Name)
	suite.Require().Equal(result["verify_trace"].(map[string]interface{})["trace"].(map[string]interface{})["counterparty_name"].(string), chainA.Name)
	suite.Require().Equal(result["verify_trace"].(map[string]interface{})["trace"].(map[string]interface{})["channel"].(string), primary_channels[chainA.Name].(string))

}