package tests

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	chainModels "github.com/allinbits/demeris-api-server/api/chains"
)

const (
	chainSupplyEndpoint = "chain/%s/supply"
	supplyKey           = "supply"
)

func (suite *testCtx) TestChainSupply() {
	for _, ch := range suite.Chains {
		suite.Run(ch.ChainName, func() {
			if ch.ChainName == "crypto-org" {
				suite.T().Skip("skip: crypto-org, sdk-service replies with Status:Unavailable")
			}

			// arrange
			url := suite.Client.BuildUrl(chainSupplyEndpoint, ch.ChainName)
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

				var coins chainModels.SupplyResponse
				err = json.Unmarshal(data, &coins)
				suite.Require().NoError(err)

				//check if the repsonse is empty
				suite.Require().NotEmpty(coins)
			}
		})
	}
}
