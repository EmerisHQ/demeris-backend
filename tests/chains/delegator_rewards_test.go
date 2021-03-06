package tests

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	api "github.com/emerishq/demeris-api-server/api/account"
	chainclient "github.com/emerishq/demeris-backend/chainclient"
)

const (
	delegatorRewardsEndpoint = "account/%v/delegatorrewards/%s"
)

func (suite *testCtx) TestDelegatorRewards() {
	for _, ch := range suite.clientChains {
		suite.Run(ch.ChainName, func() {
			cli, err := chainclient.GetClient(suite.Env, ch.ChainName, ch, suite.T().TempDir())
			suite.Require().NoError(err)
			address, err := cli.GetAccAddress(ch.Key)
			suite.Require().NoError(err)

			// arrange
			url := suite.Client.BuildUrl(delegatorRewardsEndpoint, hex.EncodeToString(address), ch.ChainName)
			// act
			resp, err := suite.Client.Get(url)
			suite.Require().NoError(err)
			// assert
			if !cli.Enabled {
				suite.Require().Equal(http.StatusBadRequest, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.ChainName, resp.StatusCode))

			} else {
				suite.Require().Equal(http.StatusOK, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.ChainName, resp.StatusCode))

				data, err := ioutil.ReadAll(resp.Body)
				suite.Require().NoError(err)

				var rewards api.DelegatorRewardsResponse
				suite.Require().NoError(json.Unmarshal(data, &rewards))
			}

			err = resp.Body.Close()
			suite.Require().NoError(err)
		})
	}
}
