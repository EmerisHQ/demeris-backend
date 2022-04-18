package tests

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	chainModels "github.com/allinbits/demeris-api-server/api/chains"
	"github.com/allinbits/demeris-backend-models/cns"
)

const chainBech32Endpoint = "chain/%s/bech32"

func (suite *testCtx) TestChainBech32() {
	for _, ch := range suite.Chains {
		suite.T().Run(ch.Name, func(t *testing.T) {
			// arrange
			url := suite.Client.BuildUrl(chainBech32Endpoint, ch.Name)
			// act
			resp, err := suite.Client.Get(url)
			suite.Require().NoError(err)

			// assert
			if !ch.Enabled {
				suite.Require().Equal(http.StatusBadRequest, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))
			} else {
				suite.Require().Equal(http.StatusOK, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))

				data, err := ioutil.ReadAll(resp.Body)
				suite.Require().NoError(err)

				err = resp.Body.Close()
				suite.Require().NoError(err)

				var bech32 chainModels.Bech32ConfigResponse
				err = json.Unmarshal(data, &bech32)
				suite.Require().NoError(err)

				suite.Require().NotEmpty(bech32)

				var expectedNodeInfo cns.Chain
				err = json.Unmarshal(ch.Payload, &expectedNodeInfo)
				suite.Require().NoError(err)

				suite.Require().Equal(expectedNodeInfo.NodeInfo.Bech32Config, bech32.Bech32Config)
			}
		})
	}
}
