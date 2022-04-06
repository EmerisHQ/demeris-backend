package tests

import (
	"fmt"
	"net/http"
	"testing"

	utils "github.com/allinbits/demeris-backend/test_utils"
)

const chainFeeAddressEndpoint = "chain/%s/fee/address"

func (suite *testCtx) TestChainFeeAddress() {
	for _, ch := range suite.Chains {
		suite.T().Run(ch.ChainName, func(t *testing.T) {
			// arrange
			url := suite.Client.BuildUrl(chainFeeAddressEndpoint, ch.ChainName)
			// act
			resp, err := suite.Client.Get(url)
			suite.Require().NoError(err)

			// assert
			if !ch.Enabled {
				suite.Require().Equal(http.StatusBadRequest, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.ChainName, resp.StatusCode))
			} else {
				suite.Require().Equal(http.StatusOK, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.ChainName, resp.StatusCode))

				// var payload map[string]interface{}
				// err := json.Unmarshal(ch.Payload, &payload)
				// suite.Require().NoError(err)

				var respValues map[string]interface{}
				utils.RespBodyToMap(resp.Body, &respValues, t)

				err = resp.Body.Close()
				suite.Require().NoError(err)

				suite.Require().Equal(ch.DemerisAddresses, respValues["fee_address"])
			}
		})
	}
}
