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

	for _, ch := range suite.Chains {
		suite.T().Run(ch.Name, func(t *testing.T) {

			// arrange
			url := suite.Client.BuildUrl(mintInflationEndpoint, ch.Name)
			// act
			resp, err := suite.Client.Get(url)
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
