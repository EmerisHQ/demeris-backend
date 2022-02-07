package tests

import (
	"fmt"
	"net/http"
	"testing"

	utils "github.com/allinbits/demeris-backend/test_utils"
)

const (
	mintInflationEndpoint = "chain/%s/mint/inflation"
	inflationKey          = "inflation"
)

func (suite *testCtx) TestMintInflation() {
	suite.T().Parallel()

	for _, ch := range suite.chains {
		suite.T().Run(ch.Name, func(t *testing.T) {

			// arrange
			url := fmt.Sprintf(baseUrl+mintInflationEndpoint, suite.emIngress.Protocol, suite.emIngress.Host, suite.emIngress.APIServerPath, ch.Name)
			// act
			resp, err := suite.client.Get(url)
			suite.NoError(err)

			defer resp.Body.Close()

			// assert
			if !ch.Enabled {
				suite.Equal(http.StatusBadRequest, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))
			} else {
				suite.Equal(http.StatusOK, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))

				var respValues map[string]interface{}
				utils.RespBodyToMap(resp.Body, &respValues, t)

				//expect a numeric value
				inflation := respValues[inflationKey]
				suite.NotZero(inflation)
			}
		})
	}
}
