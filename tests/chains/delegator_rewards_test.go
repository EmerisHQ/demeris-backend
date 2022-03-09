package tests

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	api "github.com/allinbits/demeris-api-server/api/account"
	chainClient "github.com/allinbits/demeris-backend/chain_client"
)

const (
	delegatorRewardsEndpoint = "account/%v/delegatorrewards/%s"
)

func (suite *testCtx) TestDelegatorRewards() {
	for _, ch := range suite.clientChains {
		suite.Run(ch.Name, func() {
			var cc chainClient.ChainClient
			err := json.Unmarshal(ch.Payload, &cc)
			suite.Require().NoError(err)

			cli := chainClient.GetClient(suite.T(), suite.Env, ch.Name, cc)
			address, err := cli.GetHexAddress(cc.Key)
			suite.Require().NoError(err)

			// arrange
			url := suite.Client.BuildUrl(delegatorRewardsEndpoint, hex.EncodeToString(address), ch.Name)
			// act
			resp, err := suite.Client.Get(url)
			suite.Require().NoError(err)
			// assert
			if !cli.Enabled {
				suite.Require().Equal(http.StatusBadRequest, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))

			} else {
				suite.Require().Equal(http.StatusOK, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))

				data, err := ioutil.ReadAll(resp.Body)
				suite.Require().NoError(err)

				var rewards api.DelegatorRewardsResponse
				suite.Require().NoError(json.Unmarshal(data, &rewards))
				suite.Require().NotEmpty(rewards.Rewards)
			}

			err = resp.Body.Close()
			suite.Require().NoError(err)
		})
	}
}
