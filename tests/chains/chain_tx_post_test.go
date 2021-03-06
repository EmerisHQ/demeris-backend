package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	txModels "github.com/emerishq/demeris-api-server/api/tx"
	chainclient "github.com/emerishq/demeris-backend/chainclient"
	utils "github.com/emerishq/demeris-backend/test_utils"
)

const (
	postTxEndpoint = "tx/%s"
)

func (suite *testCtx) TestTxPostEndpoint() {
	suite.T().Skip("skip: api-server failing with 'transaction relaying error: code 32, account sequence mismatch, expected 9, got 0: incorrect account sequence'")

	for _, ch := range suite.clientChains {
		suite.Run(ch.ChainName, func() {
			cli, err := chainclient.GetClient(suite.Env, ch.ChainName, ch, suite.T().TempDir())
			suite.Require().NoError(err)

			// assert
			if !cli.Enabled {
				return
			}

			// create valid tx bytes
			account, err := cli.AccountGet(ch.Key)
			suite.Require().NoError(err)

			fromAddr, err := sdk.AccAddressFromBech32(account.Address)
			suite.Require().NoError(err)

			account2, err := cli.AccountCreate("key2", "", cli.HDPath)
			suite.Require().NoError(err)

			toAddr, err := sdk.AccAddressFromBech32(account2.Address)
			suite.Require().NoError(err)

			// check balance
			balance, err := cli.GetAccountBalances(account.Address, cli.Denom)
			suite.Require().NoError(err)
			suite.Require().True(balance.Amount.GT(sdk.NewInt(0)), "not enough balance in given account to perform tx")

			// perform bank send tx
			msg := banktypes.NewMsgSend(fromAddr, toAddr, sdk.NewCoins(sdk.NewCoin(cli.Denom, sdk.NewInt(10))))

			txBytes, err := cli.SignTx(ch.Key, fromAddr, cli.GetContext(), msg)
			suite.Require().NoError(err)

			postBytes, err := json.Marshal(txModels.TxRequest{
				Owner:   account.Address,
				TxBytes: txBytes,
			})
			suite.Require().NoError(err)

			url := suite.Client.BuildUrl(postTxEndpoint, ch.ChainName)

			resp, err := suite.Client.Post(url, "application/json", bytes.NewBuffer(postBytes))
			suite.Require().NoError(err)

			suite.Require().Equal(http.StatusOK, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.ChainName, resp.StatusCode))

			data, err := ioutil.ReadAll(resp.Body)
			suite.Require().NoError(err)

			var txRes txModels.TxResponse
			suite.Require().NoError(json.Unmarshal(data, &txRes))
			suite.Require().NotEmpty(txRes)

			hash := txRes.Ticket

			err = resp.Body.Close()
			suite.Require().NoError(err)

			var nodeRes *sdktx.GetTxResponse
			err = utils.RetryOnError(func() error {
				var innerErr error
				nodeRes, innerErr = sdktx.NewServiceClient(cli.GetContext()).GetTx(context.Background(), &sdktx.GetTxRequest{Hash: hash})
				return innerErr
			}, 500*time.Millisecond, 20)
			suite.Require().NoError(err)

			suite.Require().NotEmpty(nodeRes.TxResponse)
			suite.Require().Equal(hash, nodeRes.TxResponse.TxHash)
			suite.Require().Equal(uint32(0), nodeRes.TxResponse.Code)
		})
	}
}
