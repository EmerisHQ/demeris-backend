package tests

import (
	"encoding/json"
	"net/http"

	utils "github.com/emerishq/demeris-backend/test_utils"
)

const chainsFeeAddressesEndpoint = "chains/fee/addresses"

func (suite *testCtx) TestChainsFeeAddresses() {
	// arrange
	url := suite.Client.BuildUrl(chainsFeeAddressesEndpoint)
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
			expValues["fee_addresses"] = append(expValues["fee_addresses"], map[string]interface{}{
				"chain_name":  ch.ChainName,
				"fee_address": ch.DemerisAddresses,
			})
		}
	}

	expValuesData, err := json.Marshal(expValues)
	suite.Require().NoError(err)

	var expValuesInterface map[string]interface{}
	err = json.Unmarshal(expValuesData, &expValuesInterface)
	suite.Require().NoError(err)

	suite.Require().ElementsMatch(expValuesInterface["fee_addresses"], respValues["fee_addresses"])
}
