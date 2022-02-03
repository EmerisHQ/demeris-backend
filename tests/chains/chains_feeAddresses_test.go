package tests

import (
	"encoding/json"
	"fmt"
	"net/http"

	utils "github.com/allinbits/demeris-backend/test_utils"
)

const chainsFeeAddressesEndpoint = "chains/fee/addresses"

func (suite *testCtx) TestChainsFeeAddresses() {
	suite.T().Parallel()

	// arrange
	url := fmt.Sprintf(baseUrl+chainsFeeAddressesEndpoint, suite.emIngress.Protocol, suite.emIngress.Host, suite.emIngress.APIServerPath)
	// act
	resp, err := suite.client.Get(url)
	suite.NoError(err)

	suite.Equal(http.StatusOK, resp.StatusCode)

	var respValues map[string]interface{}
	utils.RespBodyToMap(resp.Body, &respValues, suite.T())

	err = resp.Body.Close()
	suite.NoError(err)

	expValues := make(map[string][]map[string]interface{}, 0)
	for _, ch := range suite.chains {
		if ch.Enabled {
			var payload map[string]interface{}
			err := json.Unmarshal(ch.Payload, &payload)
			suite.NoError(err)

			expValues["fee_addresses"] = append(expValues["fee_addresses"], map[string]interface{}{
				"chain_name":  ch.Name,
				"fee_address": payload["demeris_addresses"],
			})
		}
	}

	expValuesData, err := json.Marshal(expValues)
	suite.NoError(err)

	var expValuesInterface map[string]interface{}
	err = json.Unmarshal(expValuesData, &expValuesInterface)
	suite.NoError(err)

	suite.ElementsMatch(expValuesInterface["fee_addresses"], respValues["fee_addresses"])
}
