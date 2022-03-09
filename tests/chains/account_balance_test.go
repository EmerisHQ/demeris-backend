package tests

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	models "github.com/allinbits/demeris-api-server/api/account"
	chainClient "github.com/allinbits/demeris-backend/chain_client"
)

const (
	getBalanceEndpoint = "/account/%v/balance"
)

func (suite *testCtx) TestGetBalanceOfAnyAccount() {
	for _, ch := range suite.clientChains {
		suite.Run(ch.Name, func() {
			var cc chainClient.ChainClient
			err := json.Unmarshal(ch.Payload, &cc)
			suite.Require().NoError(err)
			cli := chainClient.GetClient(suite.T(), suite.Env, ch.Name, cc)
			suite.Require().NotNil(cli)

			hexAddress, err := cli.GetHexAddress(cc.Key)
			suite.Require().NoError(err)

			url := suite.Client.BuildUrl(getBalanceEndpoint, hex.EncodeToString(hexAddress))
			suite.T().Log("url", url)
			// act
			resp, err := suite.Client.Get(url)
			suite.Require().NoError(err)

			suite.Require().Equal(http.StatusOK, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))

			data, err := ioutil.ReadAll(resp.Body)
			suite.Require().NoError(err)

			err = resp.Body.Close()
			suite.Require().NoError(err)

			if !cli.Enabled {
				return
			}

			var balances models.BalancesResponse
			suite.Require().NoError(json.Unmarshal(data, &balances))
			suite.Require().NotEmpty(balances.Balances)

			for _, v := range balances.Balances {
				if v.BaseDenom == cli.Denom && cli.Enabled {
					bal, err := cli.GetAccountBalances(hexAddress.String(), cli.Denom)
					suite.Require().NoError(err)
					suite.Require().Equal(bal.String(), v.Amount)
				}
			}
		})
	}
}
