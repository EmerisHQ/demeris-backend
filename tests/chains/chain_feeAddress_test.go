package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	utils "github.com/allinbits/demeris-backend/test_utils"
)

const chainFeeAddressEndpoint = "chain/%s/fee/address"

func (suite *testCtx) TestChainFeeAddress() {
	suite.T().Parallel()

	for _, ch := range suite.Chains {
		suite.T().Run(ch.Name, func(t *testing.T) {
			// arrange
			url := suite.Client.BuildUrl(chainFeeAddressEndpoint, ch.Name)
			// act
			resp, err := suite.Client.Get(url)
			suite.NoError(err)

			// assert
			if !ch.Enabled {
				suite.Equal(http.StatusBadRequest, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))
			} else {
				suite.Equal(http.StatusOK, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))

				var payload map[string]interface{}
				err := json.Unmarshal(ch.Payload, &payload)
				suite.NoError(err)

				var respValues map[string]interface{}
				utils.RespBodyToMap(resp.Body, &respValues, t)

				err = resp.Body.Close()
				suite.NoError(err)

				suite.Equal(payload["demeris_addresses"], respValues["fee_address"])
			}
		})
	}
}
