package tests

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	chainModels "github.com/emerishq/demeris-api-server/api/chains"
)

const chainsStatusEndpoint = "chains/status"

func (suite *testCtx) TestChainsStatus() {
	// arrange
	url := suite.Client.BuildUrl(chainsStatusEndpoint)
	// act
	resp, err := suite.Client.Get(url)
	suite.Require().NoError(err)

	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	suite.Require().NoError(err)

	var chStatus chainModels.ChainsStatusesResponse
	err = json.Unmarshal(data, &chStatus)
	suite.Require().NoError(err)

	for _, ch := range suite.Chains {
		suite.T().Run(ch.ChainName, func(t *testing.T) {
			// assert
			if ch.Enabled {
				suite.Require().Equal(http.StatusOK, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.ChainName, resp.StatusCode))

				// check the chain status
				suite.Require().Equal(true, chStatus.Chains[ch.ChainName].Online)
			}
		})
	}
}
