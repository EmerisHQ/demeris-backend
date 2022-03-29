package tests

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	chainModels "github.com/allinbits/demeris-api-server/api/chains"
)

const chainFeeTokenEndpoint = "chain/%s/fee/token"

func (suite *testCtx) TestChainFeeToken() {
	for _, ch := range suite.Chains {
		suite.T().Run(ch.Name, func(t *testing.T) {
			// arrange
			url := suite.Client.BuildUrl(chainFeeTokenEndpoint, ch.Name)
			// act
			resp, err := suite.Client.Get(url)
			suite.NoError(err)

			// assert
			if !ch.Enabled {
				suite.Equal(http.StatusBadRequest, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))
			} else {
				suite.Equal(http.StatusOK, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))

				data, err := ioutil.ReadAll(resp.Body)
				suite.NoError(err)

				var denoms chainModels.FeeResponse
				err = json.Unmarshal(data, &denoms)
				suite.NoError(err)

				suite.NotEmpty(denoms)

				var expectedDenoms chainModels.FeeResponse
				err = json.Unmarshal(ch.Payload, &expectedDenoms)
				suite.NoError(err)
				suite.NotNil(expectedDenoms)

				suite.ElementsMatch(denoms, expectedDenoms)
			}
		})
	}
}
