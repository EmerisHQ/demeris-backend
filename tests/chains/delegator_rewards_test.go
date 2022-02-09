package tests

import (
	"encoding/json"
	"fmt"
	"net/http"

	chainClient "github.com/allinbits/demeris-backend/chain_client"
)

const (
	chainTxsEndpoint = "account/%v/delegatorrewards/%s"
	randomTxHash     = "56FF608A76A01D9178039D17949F53ED8E3969752D546E5474605A67B13A42A0"
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
			url := fmt.Sprintf(baseUrl+chainTxsEndpoint, suite.emIngress.Protocol, suite.emIngress.Host, suite.emIngress.APIServerPath,
				address, ch.Name)
			// act
			resp, err := suite.client.Get(url)
			suite.Require().NoError(err)
			// assert
			if !cli.Enabled {
				suite.Require().Equal(http.StatusBadRequest, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))

			} else {
				suite.Require().Equal(http.StatusOK, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))
			}
			err = resp.Body.Close()
			suite.Require().NoError(err)

			//TODO: make DelegatorRewardsResponse public in api-server
		})
	}
}
