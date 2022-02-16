package tests

import (
	"encoding/json"
	"fmt"
	"net/http"

	utils "github.com/allinbits/demeris-backend/test_utils"
)

const chainsEndpoint = "chains"

func (suite *testCtx) TestChainsData() {
	suite.T().Parallel()

	// arrange
	url := suite.Client.BuildUrl(chainsEndpoint)
	// act
	resp, err := suite.Client.Get(url)
	suite.NoError(err)

	suite.Equal(http.StatusOK, resp.StatusCode)

	var respValues map[string]interface{}
	utils.RespBodyToMap(resp.Body, &respValues, suite.T())

	err = resp.Body.Close()
	suite.NoError(err)

	expValues := make(map[string][]map[string]interface{}, 0)
	for _, ch := range suite.Chains {
		if ch.Enabled {
			chainUrl := suite.Client.BuildUrl("chain/%s", ch.Name)
			chainResp, err := suite.Client.Get(chainUrl)
			suite.NoError(err)

			suite.Equal(http.StatusOK, chainResp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, chainResp.StatusCode))

			var payload map[string]interface{}
			err = json.Unmarshal(ch.Payload, &payload)
			suite.NoError(err)

			expValues["chains"] = append(expValues["chains"], map[string]interface{}{
				"chain_name":   ch.Name,
				"display_name": payload["display_name"],
				"logo":         payload["logo"],
			})
		}
	}

	expValuesData, err := json.Marshal(expValues)
	suite.NoError(err)

	var expValuesInterface map[string]interface{}
	err = json.Unmarshal(expValuesData, &expValuesInterface)
	suite.NoError(err)

	suite.ElementsMatch(expValuesInterface["chains"], respValues["chains"])
}
