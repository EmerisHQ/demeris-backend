package tests

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/allinbits/demeris-backend-models/cns"
	chainClient "github.com/allinbits/demeris-backend/chain_client"
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
	var enabledChains []cns.Chain
	for _, chain := range suite.Chains {
		if chain.Enabled {
			enabledChains = append(enabledChains, chain)
		}
	}
	suite.Require().Greater(len(enabledChains), 1, "Need atleast 2 enabled chains to perform IBC transaction")

	// pick 2 random chains
	var chainA, chainB cns.Chain
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
	var ccA, ccB chainClient.ChainClient
	for _, ch := range suite.clientChains {
		if ch.ChainName == chainA.ChainName {
			ccA = ch
		} else if ch.ChainName == chainB.ChainName {
			ccB = ch
		}
	}
	cliB, err := chainClient.GetClient(suite.Env, chainB.ChainName, ccB, suite.T().TempDir())
	suite.NoError(err)
	suite.Require().NotNil(cliB)
	recAccount, err := cliB.AccountGet(ccB.Key)
	suite.Require().NoError(err)

	cliA, err := chainClient.GetClient(suite.Env, chainA.ChainName, ccA, suite.T().TempDir())
	suite.NoError(err)

	suite.Require().NotNil(cliA)
	send_account, err := cliA.AccountGet(ccA.Key)
	suite.Require().NoError(err)
	fromAddr, err := sdk.AccAddressFromBech32(send_account.Address)
	suite.Require().NoError(err)

	primary_channels := chainA.PrimaryChannel

	// check balance for account A
	accABalance, err := cliA.GetAccountBalances(send_account.Address, chainA.Denoms[0].DisplayName)
	suite.Require().NoError(err)
	suite.Require().Greater(accABalance.Amount.BigInt().Uint64(), uint64(100), "Not enough balance to make an IBC transaction")

	token := sdk.Coin{
		Denom:  chainA.Denoms[0].DisplayName,
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
	ibcDenom := strings.ToUpper(fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("transfer/%s/%s", primary_channels[ccB.ChainName], chainA.Denoms[0].DisplayName)))))

	// get account B balance before IBC transaction
	prevBalance, err := cliB.GetAccountBalances(recAccount.Address, fmt.Sprintf("ibc/%s", ibcDenom))
	suite.Require().NoError(err)

	// build and broadcast ibc transfer message
	msg := ibctransfertypes.NewMsgTransfer("transfer", primary_channels[ccB.ChainName], token, fromAddr, recAccount.Address, timeoutHeight, 0)

	_, err = cliA.Broadcast(ccA.Key, cliA.GetContext(), msg)
	suite.Require().NoError(err)

	time.Sleep(time.Second * 10)

	// get account B balance after IBC transaction
	postBalance, err := cliB.GetAccountBalances(recAccount.Address, fmt.Sprintf("ibc/%s", ibcDenom))
	suite.Require().NoError(err)

	// check updated balance
	suite.Require().Equal(uint64(100), postBalance.Amount.BigInt().Uint64()-prevBalance.Amount.BigInt().Uint64())

	url := suite.Client.BuildUrl(verifyTraceEndpoint, chainB.ChainName, ibcDenom)
	// act
	respTrace, err := suite.Client.Get(url)
	suite.Require().NoError(err)
	suite.Require().Equal(http.StatusOK, respTrace.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", chainB.ChainName, resp.StatusCode))

	body, err := ioutil.ReadAll(respTrace.Body)
	suite.Require().NoError(err)
	var result map[string]interface{}
	suite.Require().NoError(json.Unmarshal(body, &result))

	primary_channels = chainB.PrimaryChannel
	fmt.Println(result)
	suite.Require().Equal(result["verify_trace"].(map[string]interface{})["base_denom"].(string), chainA.Denoms[0].DisplayName)
	suite.Require().Equal(result["verify_trace"].(map[string]interface{})["path"].(string), fmt.Sprintf("transfer/%s", primary_channels[chainA.ChainName]))
	suite.Require().Equal(result["verify_trace"].(map[string]interface{})["verified"].(bool), chainA.Denoms[0].Verified)
	suite.Require().Equal(result["verify_trace"].(map[string]interface{})["trace"].(map[string]interface{})["chain_name"].(string), chainB.ChainName)
	suite.Require().Equal(result["verify_trace"].(map[string]interface{})["trace"].(map[string]interface{})["counterparty_name"].(string), chainA.ChainName)
	suite.Require().Equal(result["verify_trace"].(map[string]interface{})["trace"].(map[string]interface{})["channel"].(string), primary_channels[chainA.ChainName])

}
