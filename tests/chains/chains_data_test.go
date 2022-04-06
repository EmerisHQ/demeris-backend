package tests

import (
	"encoding/json"
	"fmt"
	"net/http"

	utils "github.com/allinbits/demeris-backend/test_utils"
)

const chainsEndpoint = "chains"

func (suite *testCtx) TestChainsData() {
	// arrange
	url := suite.Client.BuildUrl(chainsEndpoint)
	// act
	resp, err := suite.Client.Get(url)
	suite.Require().NoError(err)

	suite.Require().Equal(http.StatusOK, resp.StatusCode)

	var respValues map[string]interface{}
	utils.RespBodyToMap(resp.Body, &respValues, suite.T())

	err = resp.Body.Close()
	suite.Require().NoError(err)

	expValues := make(map[string][]map[string]interface{}, 0)
	for _, ch := range suite.Chains {
		if ch.Enabled {
			chainUrl := suite.Client.BuildUrl("chain/%s", ch.ChainName)
			chainResp, err := suite.Client.Get(chainUrl)
			suite.Require().NoError(err)

			suite.Require().Equal(http.StatusOK, chainResp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.ChainName, chainResp.StatusCode))

			// var payload map[string]interface{}
			// err = json.Unmarshal(ch.Payload, &payload)
			// suite.Require().NoError(err)

			expValues["chains"] = append(expValues["chains"], map[string]interface{}{
				"chain_name":   ch.ChainName,
				"display_name": ch.DisplayName,
				"logo":         ch.Logo,
			})
		}
	}

	expValuesData, err := json.Marshal(expValues)
	suite.Require().NoError(err)

	var expValuesInterface map[string]interface{}
	err = json.Unmarshal(expValuesData, &expValuesInterface)
	suite.Require().NoError(err)

	suite.Require().ElementsMatch(expValuesInterface["chains"], respValues["chains"])
}
