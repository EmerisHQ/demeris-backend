package tests

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctransfertypes "github.com/cosmos/cosmos-sdk/x/ibc/applications/transfer/types"
	ibcclienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	"github.com/emerishq/demeris-backend-models/cns"
	chainclient "github.com/emerishq/demeris-backend/chainclient"
)

const (
	getDestTxnEndpoint = "tx/%s/%s/%s"
)

var chainsFilter = map[string]bool{
	"akash":      true,
	"cosmos-hub": true,
	"terra":      false,
}

func (suite *testCtx) TestGetDestTxn() {

	// filter enabled chains
	var enabledChains []cns.Chain
	for _, chain := range suite.Chains {
		if chain.Enabled && chainsFilter[chain.ChainName] {
			enabledChains = append(enabledChains, chain)
		}
	}
	suite.Require().Greater(len(enabledChains), 1, "Need atleast 2 enabled chains to perform IBC transaction")

	// pick 2 random chains
	var chainA, chainB cns.Chain
	for {
		a := rand.Intn(len(enabledChains))
		b := rand.Intn(len(enabledChains))
		if a != b {
			chainA = enabledChains[a]
			chainB = enabledChains[b]
			break
		}
	}

	// create clients and accounts for above picked chains
	var ccA, ccB chainclient.ChainClient
	for _, ch := range suite.clientChains {
		if ch.ChainName == chainA.ChainName {
			ccA = ch
		} else if ch.ChainName == chainB.ChainName {
			ccB = ch
		}
	}
	cliB, err := chainclient.GetClient(suite.Env, chainB.ChainName, ccB, suite.T().TempDir())
	suite.Require().NoError(err)
	rec_account, err := cliB.AccountGet(ccB.Key)
	suite.Require().NoError(err)

	cliA, err := chainclient.GetClient(suite.Env, chainA.ChainName, ccA, suite.T().TempDir())
	suite.Require().NoError(err)
	send_account, err := cliA.AccountGet(ccA.Key)
	suite.Require().NoError(err)
	fromAddr, err := cliA.GetAccAddress(ccA.Key)
	suite.Require().NoError(err)

	primary_channels := chainA.PrimaryChannel

	// // get chainA denom
	denom := chainA.Denoms[0].DisplayName

	// check balance for account A
	accABalance, err := cliA.GetAccountBalances(send_account.Address, denom)
	suite.Require().NoError(err)
	suite.Require().Greater(accABalance.Amount.BigInt().Uint64(), uint64(100), "Not enough balance to make an IBC transaction")

	token := sdk.Coin{
		Denom:  denom,
		Amount: sdk.NewInt(100),
	}

	// get current block height and set timeout height
	resp, err := http.Get(ccB.RPC + "/block?height")
	suite.Require().NoError(err)

	bz, err := ioutil.ReadAll(resp.Body)
	suite.Require().NoError(err)
	defer suite.Require().NoError(resp.Body.Close())

	var blockData map[string]interface{}
	suite.Require().NoError(json.Unmarshal(bz, &blockData))

	heightFromResp := blockData["result"].(map[string]interface{})["block"].(map[string]interface{})["header"].(map[string]interface{})["height"].(string)
	height, err := strconv.Atoi(heightFromResp)
	suite.Require().NoError(err)

	timeoutHeight := ibcclienttypes.Height{
		RevisionNumber: 0,
		RevisionHeight: uint64(height + 100000),
	}

	// build IBC denom hash
	ibcDenom := strings.ToUpper(fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("transfer/%s/%s", primary_channels[ccB.ChainName], denom)))))

	// get account B balance before IBC transaction
	prevBalance, err := cliB.GetAccountBalances(rec_account.Address, fmt.Sprintf("ibc/%s", ibcDenom))
	suite.Require().NoError(err)

	// build and broadcast ibc transfer message
	msg := ibctransfertypes.NewMsgTransfer("transfer", primary_channels[ccB.ChainName], token, fromAddr, rec_account.Address, timeoutHeight, 0)

	txRes, err := cliA.Broadcast(ccA.Key, fromAddr, cliA.GetContext(), msg)
	suite.Require().NoError(err)

	time.Sleep(time.Second * 8)

	// get account B balance after IBC transaction
	postBalance, err := cliB.GetAccountBalances(rec_account.Address, fmt.Sprintf("ibc/%s", ibcDenom))
	suite.Require().NoError(err)

	// check updated balance
	suite.Require().Equal(uint64(100), postBalance.Amount.BigInt().Uint64()-prevBalance.Amount.BigInt().Uint64())

	// get dest tx ID
	url := suite.Client.BuildUrl(getDestTxnEndpoint, chainA.ChainName, chainB.ChainName, txRes.TxHash)
	respDestTx, err := suite.Client.Get(url)
	suite.Require().NoError(err)
	suite.Require().Equal(http.StatusOK, respDestTx.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", chainB.ChainName, resp.StatusCode))

	destTxnBody, err := ioutil.ReadAll(respDestTx.Body)
	suite.Require().NoError(err)
	defer suite.Require().NoError(respDestTx.Body.Close())

	var resultDestTx map[string]interface{}
	suite.Require().NoError(json.Unmarshal(destTxnBody, &resultDestTx))

	// get txBody chainB
	url = suite.Client.BuildUrl(chainTxsEndpoint, chainB.ChainName, resultDestTx["tx_hash"].(string))
	respTxnChainB, err := suite.Client.Get(url)
	suite.Require().NoError(err)
	suite.Require().Equal(http.StatusOK, respTxnChainB.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", chainB.ChainName, resp.StatusCode))

	txnBodyChainB, err := ioutil.ReadAll(respTxnChainB.Body)
	suite.Require().NoError(err)
	defer suite.Require().NoError(respTxnChainB.Body.Close())

	var resultTxnChainB map[string]interface{}
	suite.Require().NoError(json.Unmarshal(txnBodyChainB, &resultTxnChainB))
	suite.Require().NotNil(resultTxnChainB["tx"])
}
