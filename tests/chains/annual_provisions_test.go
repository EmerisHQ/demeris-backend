package tests

import (
	"fmt"
	"net/http"

	utils "github.com/emerishq/demeris-backend/test_utils"
)

const (
	mintAnnualProvisionsEndpoint = "chain/%s/mint/annual_provisions"
	annualProvisionKey           = "annual_provisions"
)

func (suite *testCtx) TestAnnualProvisions() {
	for _, ch := range suite.Chains {
		suite.Run(ch.ChainName, func() {
			if ch.ChainName == "osmosis" || ch.ChainName == "crypto-org" {
				suite.T().Skipf("skip(%s): api-server failing", ch.ChainName)
			}

			// arrange
			url := suite.Client.BuildUrl(mintAnnualProvisionsEndpoint, ch.ChainName)
			// act
			resp, err := suite.Client.Get(url)
			suite.Require().NoError(err)

			defer resp.Body.Close()

			// assert
			if !ch.Enabled {
				suite.Require().Equal(http.StatusBadRequest, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.ChainName, resp.StatusCode))
			} else {
				suite.Require().Equal(http.StatusOK, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.ChainName, resp.StatusCode))

				var respValues map[string]interface{}
				utils.RespBodyToMap(resp.Body, &respValues, suite.T())

				//expect a non empty data
				provisions := respValues[annualProvisionKey]
				suite.Require().NotEmpty(provisions)
			}
		})
	}
}
