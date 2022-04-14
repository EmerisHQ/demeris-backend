package tests

import (
	"fmt"
	"net/http"
	"time"

	chainclient "github.com/allinbits/demeris-backend/chainclient"
	utils "github.com/allinbits/demeris-backend/test_utils"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

const (
	chainTxsEndpoint = "chain/%s/txs/%s"

	// will use this only for disabled chains
	randomTxHash = "56FF608A76A01D9178039D17949F53ED8E3969752D546E5474605A67B13A42A0"
)

func (suite *testCtx) TestTxsEndpoint() {
	for _, ch := range suite.clientChains {
		suite.Run(ch.ChainName, func() {
			cli, err := chainclient.GetClient(suite.Env, ch.ChainName, ch, suite.T().TempDir())
			suite.Require().NoError(err)
			// assert
			if !cli.Enabled {
				// arrange
				url := suite.Client.BuildUrl(chainTxsEndpoint, ch.ChainName, randomTxHash)
				// act
				resp, err := suite.Client.Get(url)
				suite.Require().NoError(err)
				suite.Require().Equal(http.StatusBadRequest, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.ChainName, resp.StatusCode))
				err = resp.Body.Close()
				suite.Require().NoError(err)
			} else {
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
				msg := banktypes.NewMsgSend(fromAddr, toAddr, sdk.NewCoins(sdk.NewCoin(cli.Denom, sdk.NewInt(100))))
				txRes, err := cli.Broadcast(ch.Key, cli.GetContext(), msg)
				suite.Require().NoError(err)

				hash := txRes.TxHash

				time.Sleep(time.Second * 8)

				// arrange
				url := suite.Client.BuildUrl(chainTxsEndpoint, ch.ChainName, hash)
				// act
				resp, err := suite.Client.Get(url)
				suite.Require().NoError(err)
				suite.Require().Equal(http.StatusOK, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.ChainName, resp.StatusCode))

				var respValues map[string]interface{}
				utils.RespBodyToMap(resp.Body, &respValues, suite.T())

				err = resp.Body.Close()
				suite.Require().NoError(err)

				txhashMap := respValues["tx_response"].(map[string]interface{})
				suite.Require().NotEmpty(txhashMap)
				suite.Require().Equal(hash, txhashMap["txhash"])
			}
		})
	}
}
