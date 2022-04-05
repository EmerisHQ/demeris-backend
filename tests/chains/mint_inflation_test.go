package tests

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	chainModels "github.com/allinbits/demeris-api-server/api/chains"
)

const (
	mintInflationEndpoint = "chain/%s/mint/inflation"
	inflationKey          = "inflation"
)

func (suite *testCtx) TestMintInflation() {
	for _, ch := range suite.Chains {
		suite.T().Run(ch.Name, func(t *testing.T) {

			// arrange
			url := suite.Client.BuildUrl(mintInflationEndpoint, ch.Name)
			// act
			resp, err := suite.Client.Get(url)
			suite.Require().NoError(err)

			defer resp.Body.Close()

			// assert
			if !ch.Enabled {
				suite.Require().Equal(http.StatusBadRequest, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))
			} else {
				suite.Require().Equal(http.StatusOK, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))

				data, err := ioutil.ReadAll(resp.Body)
				suite.Require().NoError(err)

				err = resp.Body.Close()
				suite.Require().NoError(err)

				var inflation chainModels.InflationResponse
				err = json.Unmarshal(data, &inflation)
				suite.Require().NoError(err)
			}
		})
	}
}
