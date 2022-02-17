package tests

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/allinbits/demeris-backend-models/api"
	chainClient "github.com/allinbits/demeris-backend/chain_client"
)

const (
	delegatorRewardsEndpoint = "account/%v/delegatorrewards/%s"
)

func (suite *testCtx) TestDelegatorRewards() {
	suite.T().Parallel()

	for _, ch := range suite.clientChains {
		suite.Run(ch.Name, func() {
			var cc chainClient.Client
			err := json.Unmarshal(ch.Payload, &cc)
			suite.Require().NoError(err)
			cli := chainClient.GetClient(suite.T(), suite.env, ch.Name, cc)
			address, err := cli.GetHexAddress(cc.Key)
			suite.Require().NoError(err)
			// arrange
			url := fmt.Sprintf(baseUrl+delegatorRewardsEndpoint, suite.emIngress.Protocol, suite.emIngress.Host, suite.emIngress.APIServerPath,
				hex.EncodeToString(address), ch.Name)
			// act
			resp, err := suite.client.Get(url)
			suite.Require().NoError(err)
			// assert
			if !cli.Enabled {
				suite.Require().Equal(http.StatusBadRequest, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))

			} else {
				suite.Require().Equal(http.StatusOK, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))

				data, err := ioutil.ReadAll(resp.Body)
				suite.Require().NoError(err)

				// TODO: modify backend-model dependency version in go.mod once api-server models are included in backend-models repo
				var rewards api.DelegatorRewardsResponse
				suite.Require().NoError(json.Unmarshal(data, &rewards))
				suite.Require().NotEmpty(rewards.Rewards)
			}

			err = resp.Body.Close()
			suite.Require().NoError(err)
		})
	}
}
