package tests

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	chainClient "github.com/allinbits/demeris-backend/chain_client"
	utils "github.com/allinbits/demeris-backend/test_utils"
)

const (
	getBalanceEndpoint = "/account/%v/balance"
)

func (suite *testCtx) TestGetBalanceOfAnyAccount() {
	suite.T().Parallel()

	for _, ch := range suite.clientChains {
		suite.Run(ch.Name, func() {
			var cc chainClient.Client
			err := json.Unmarshal(ch.Payload, &cc)
			suite.Require().NoError(err)
			cli := chainClient.GetClient(suite.T(), suite.Env, ch.Name, cc)
			suite.Require().NotNil(cli)

			hexAddress, err := cli.GetHexAddress(cc.Key)
			suite.Require().NoError(err)

			url := suite.Client.BuildUrl(getBalanceEndpoint, hex.EncodeToString(hexAddress))
			// act
			resp, err := suite.Client.Get(url)
			suite.Require().NoError(err)

			if !cli.Enabled {
				suite.Require().Equal(http.StatusBadRequest, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))
				err = resp.Body.Close()
				suite.Require().NoError(err)

				return
			}
			suite.Require().Equal(http.StatusOK, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))

			var respValues map[string]interface{}
			utils.RespBodyToMap(resp.Body, &respValues, suite.T())

			err = resp.Body.Close()
			suite.Require().NoError(err)
			suite.Require().NotNil(respValues)

			data, err := json.Marshal(respValues["balances"])
			suite.Require().NoError(err)
			suite.Require().NotNil(data)

			//TODO: Modify below code once https://github.com/allinbits/demeris-backend-models/pull/20 is merged
			// var row []apiServer.Balance
			// err = json.Unmarshal(data, &row)
			// suite.Require().NoError(err)
			// suite.Require().NotNil(row)

			// for _, v := range row {
			// 	if v.Denom == cli.Denom {
			// 		bal, err := cli.GetAccountBalances(cli.Key, cli.Denom)
			// 		suite.Require().NoError(err)
			// 		suite.Require().Equal(bal.Amount, v.Amount)
			// 	}
			// }
		})
	}
}
