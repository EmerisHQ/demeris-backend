package tests

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	chainModels "github.com/allinbits/demeris-api-server/api/chains"
)

const (
	statusEndpoint = "chain/%s/status"
	onlineKey      = "online"
)

func (suite *testCtx) TestChainStatus() {
	for _, ch := range suite.Chains {
		suite.T().Run(ch.ChainName, func(t *testing.T) {
			// arrange
			url := suite.Client.BuildUrl(statusEndpoint, ch.ChainName)
			// act
			resp, err := suite.Client.Get(url)
			suite.Require().NoError(err)

			// assert
			if !ch.Enabled {
				suite.Require().Equal(http.StatusBadRequest, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.ChainName, resp.StatusCode))
			} else {
				suite.Require().Equal(http.StatusOK, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.ChainName, resp.StatusCode))

				data, err := ioutil.ReadAll(resp.Body)
				suite.Require().NoError(err)

				err = resp.Body.Close()
				suite.Require().NoError(err)

				var status chainModels.StatusResponse
				err = json.Unmarshal(data, &status)
				suite.Require().NoError(err)

				suite.Require().Equal(true, status.Online)
			}
		})
	}
}
