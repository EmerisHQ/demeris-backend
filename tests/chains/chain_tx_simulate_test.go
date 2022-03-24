package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	txModels "github.com/allinbits/demeris-api-server/api/tx"
	chainClient "github.com/allinbits/demeris-backend/chain_client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

const (
	simulateTxEndpoint = "tx/%s/simulate"
)

func (suite *testCtx) TestTxSimulateEndpoint() {
	for _, ch := range suite.clientChains {
		suite.Run(ch.Name, func() {
			var cc chainClient.Client
			err := json.Unmarshal(ch.Payload, &cc)
			suite.Require().NoError(err)

			cli := chainClient.GetClient(suite.T(), suite.Env, ch.Name, cc)

			// assert
			if !cli.Enabled {
				return
			}

			// create valid tx bytes
			account, err := cli.AccountGet(cc.Key)
			suite.Require().NoError(err)

			fromAddr, err := sdk.AccAddressFromBech32(account.Address)
			suite.Require().NoError(err)

			account2, err := cli.AccountCreate("key2", "", cli.HDPath)
			suite.Require().NoError(err)

			toAddr, err := sdk.AccAddressFromBech32(account2.Address)
			suite.Require().NoError(err)

			// perform bank send tx
			msg := banktypes.NewMsgSend(fromAddr, toAddr, sdk.NewCoins(sdk.NewCoin(cli.Denom, sdk.NewInt(10))))

			txBytes, err := cli.SignTx(context.Background(), cc.Key, cli.GetContext(), msg)
			suite.Require().NoError(err)

			reqBytes, err := json.Marshal(txModels.TxFeeEstimateReq{
				TxBytes: txBytes,
			})
			suite.Require().NoError(err)

			url := suite.Client.BuildUrl(simulateTxEndpoint, ch.Name)

			resp, err := suite.Client.Post(url, "application/json", bytes.NewBuffer(reqBytes))
			suite.Require().NoError(err)

			suite.Require().Equal(http.StatusOK, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))

			data, err := ioutil.ReadAll(resp.Body)
			suite.Require().NoError(err)

			var feesRes txModels.TxFeeEstimateRes
			suite.Require().NoError(json.Unmarshal(data, &feesRes))

			// assert
			suite.Require().NotEmpty(feesRes)
			suite.Require().Greater(feesRes.GasUsed, uint64(0))

			if ch.Name == "terra" {
				suite.Require().NotEmpty(feesRes.Fees)
			}

			err = resp.Body.Close()
			suite.Require().NoError(err)
		})
	}
}