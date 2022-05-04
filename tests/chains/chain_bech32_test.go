package tests

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	chainModels "github.com/allinbits/demeris-api-server/api/chains"
)

const chainBech32Endpoint = "chain/%s/bech32"

func (suite *testCtx) TestChainBech32() {
	for _, ch := range suite.Chains {
		suite.Run(ch.ChainName, func() {
			// arrange
			url := suite.Client.BuildUrl(chainBech32Endpoint, ch.ChainName)
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

				var bech32 chainModels.Bech32ConfigResponse
				err = json.Unmarshal(data, &bech32)
				suite.Require().NoError(err)

				suite.Require().NotEmpty(bech32)

				suite.Require().Equal(ch.NodeInfo.Bech32Config, bech32.Bech32Config)
			}
		})
	}
}
