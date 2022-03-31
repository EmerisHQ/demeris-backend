package tests

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	chainModels "github.com/allinbits/demeris-api-server/api/chains"
)

const chainFeeEndpoint = "chain/%s/fee"

func (suite *testCtx) TestChainFee() {
	for _, ch := range suite.Chains {
		suite.T().Run(ch.Name, func(t *testing.T) {
			// arrange
			url := suite.Client.BuildUrl(chainFeeEndpoint, ch.Name)
			// act
			resp, err := suite.Client.Get(url)
			suite.Require().NoError(err)

			// assert
			if !ch.Enabled {
				suite.Require().Equal(http.StatusBadRequest, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))
			} else {
				suite.Require().Equal(http.StatusOK, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))

				data, err := ioutil.ReadAll(resp.Body)
				suite.Require().NoError(err)

				err = resp.Body.Close()
				suite.Require().NoError(err)

				var denoms chainModels.FeeResponse
				err = json.Unmarshal(data, &denoms)
				suite.Require().NoError(err)

				suite.Require().NotEmpty(denoms)

				var expectedDenoms chainModels.FeeResponse
				err = json.Unmarshal(ch.Payload, &expectedDenoms)
				suite.Require().NoError(err)

				suite.Require().Equal(expectedDenoms.Denoms, denoms.Denoms)
			}
		})
	}
}
