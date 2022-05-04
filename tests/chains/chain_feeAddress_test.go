package tests

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	chainModels "github.com/emerishq/demeris-api-server/api/chains"
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

				data, err := ioutil.ReadAll(resp.Body)
				suite.Require().NoError(err)

				err = resp.Body.Close()
				suite.Require().NoError(err)

				var feeAddresss chainModels.FeeAddressResponse
				err = json.Unmarshal(data, &feeAddresss)
				suite.Require().NoError(err)

				suite.Require().Equal(ch.DemerisAddresses, feeAddresss.FeeAddress)
			}
		})
	}
}
