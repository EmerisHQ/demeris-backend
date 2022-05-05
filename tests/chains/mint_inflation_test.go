package tests

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	chainModels "github.com/emerishq/demeris-api-server/api/chains"
)

const (
	mintInflationEndpoint = "chain/%s/mint/inflation"
	inflationKey          = "inflation"
)

func (suite *testCtx) TestMintInflation() {
	for _, ch := range suite.Chains {
		suite.Run(ch.ChainName, func() {
			if ch.ChainName == "crypto-org" {
				suite.T().Skip("skip: crypto-org, api-server returns error")
			}

			// arrange
			url := suite.Client.BuildUrl(mintInflationEndpoint, ch.ChainName)
			// act
			resp, err := suite.Client.Get(url)
			suite.Require().NoError(err)

			defer resp.Body.Close()

			// assert
			if !ch.Enabled {
				suite.Require().Equal(http.StatusBadRequest, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.ChainName, resp.StatusCode))
			} else {
				suite.Require().Equal(http.StatusOK, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.ChainName, resp.StatusCode))

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
