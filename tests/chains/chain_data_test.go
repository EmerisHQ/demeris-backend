package tests

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	chainModels "github.com/allinbits/demeris-api-server/api/chains"
)

const (
	chainEndpoint = "chain/%s"
	respChainKey  = "chain"
)

func (suite *testCtx) TestChainData() {
	for _, ch := range suite.Chains {
		suite.Run(ch.ChainName, func() {

			// arrange
			url := suite.Client.BuildUrl(chainEndpoint, ch.ChainName)
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
				suite.Require().NotNil(data)

				err = resp.Body.Close()
				suite.Require().NoError(err)

				var chain chainModels.ChainResponse
				err = json.Unmarshal(data, &chain)
				suite.Require().NoError(err)
				suite.Require().NotNil(chain)

				suite.Require().Equal(ch, chain.Chain)
			}
		})
	}
}
