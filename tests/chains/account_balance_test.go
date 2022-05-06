package tests

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	models "github.com/emerishq/demeris-api-server/api/account"
	chainclient "github.com/emerishq/demeris-backend/chainclient"
)

const (
	getBalanceEndpoint = "/account/%v/balance"
)

func (suite *testCtx) TestGetBalanceOfAnyAccount() {
	suite.T().Skip("skip: error: 'empty address string is not allowed' on sdktypes.AccAddressFromBech32(...)")

	for _, ch := range suite.clientChains {
		suite.Run(ch.ChainName, func() {
			cli, err := chainclient.GetClient(suite.Env, ch.ChainName, ch, suite.T().TempDir())
			suite.Require().NoError(err)
			suite.Require().NotNil(cli)

			accAddr, err := cli.GetAccAddress(ch.Key)
			suite.Require().NoError(err)

			url := suite.Client.BuildUrl(getBalanceEndpoint, hex.EncodeToString(accAddr))
			// act
			resp, err := suite.Client.Get(url)
			suite.Require().NoError(err)

			suite.Require().Equal(http.StatusOK, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.ChainName, resp.StatusCode))

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
					bal, err := cli.GetAccountBalances(accAddr.String(), cli.Denom)
					suite.Require().NoError(err)
					suite.Require().Equal(bal.String(), v.Amount)
				}
			}
		})
	}
}
