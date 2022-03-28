package tests

import (
	"fmt"
	"net/http"
	"testing"

	utils "github.com/allinbits/demeris-backend/test_utils"
)

const (
	mintAnnualProvisionsEndpoint = "chain/%s/mint/annual_provisions"
	annualProvisionKey           = "annual_provisions"
)

func (suite *testCtx) TestAnnualProvisions() {
	if suite.Env == "staging" {
		suite.T().Skip("skipping annual provisions")
	}
	for _, ch := range suite.Chains {
		suite.T().Run(ch.Name, func(t *testing.T) {
			// arrange
			url := suite.Client.BuildUrl(mintAnnualProvisionsEndpoint, ch.Name)
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

				//expect a non empty data
				provisions := respValues[annualProvisionKey]
				suite.NotEmpty(provisions)
			}
		})
	}
}
