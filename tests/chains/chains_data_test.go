package tests

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	chainModels "github.com/emerishq/demeris-api-server/api/chains"
	apiutils "github.com/emerishq/demeris-api-server/api/test_utils"
)

const chainsEndpoint = "chains"

func (suite *testCtx) TestChainsData() {
	suite.T().Skip("skip: payload format changed in api-server")

	// arrange
	url := suite.Client.BuildUrl(chainsEndpoint)
	// act
	resp, err := suite.Client.Get(url)
	suite.Require().NoError(err)

	suite.Require().Equal(http.StatusOK, resp.StatusCode)

	data, err := ioutil.ReadAll(resp.Body)
	suite.Require().NoError(err)

	err = resp.Body.Close()
	suite.Require().NoError(err)

	var respValues chainModels.ChainsResponse
	err = json.Unmarshal(data, &respValues)
	suite.Require().NoError(err)
	suite.Require().NotEmpty(respValues)

	for _, ch := range suite.Chains {
		if ch.Enabled {
			chainUrl := suite.Client.BuildUrl("chain/%s/status", ch.ChainName)
			statusResp, err := suite.Client.Get(chainUrl)
			suite.Require().NoError(err)

			suite.Require().Equal(http.StatusOK, statusResp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.ChainName, statusResp.StatusCode))

			statusData, err := ioutil.ReadAll(statusResp.Body)
			suite.Require().NoError(err)

			err = statusResp.Body.Close()
			suite.Require().NoError(err)

			var status chainModels.StatusResponse
			err = json.Unmarshal(statusData, &status)
			suite.Require().NoError(err)

			chainWithStatus := apiutils.ToChainWithStatus(ch, status.Online)
			suite.Require().Contains(respValues.Chains, chainWithStatus)
		}
	}
}
