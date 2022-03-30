package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/allinbits/demeris-backend-models/cns"
	utils "github.com/allinbits/demeris-backend/test_utils"
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

				var respValues map[string]interface{}
				utils.RespBodyToMap(resp.Body, &respValues, t)

				err = resp.Body.Close()
				suite.Require().NoError(err)

				data, err := json.Marshal(respValues["bech32_config"])
				suite.Require().NoError(err)

				var bech32 cns.Bech32Config
				err = json.Unmarshal(data, &bech32)
				suite.Require().NoError(err)

				suite.Require().NotEmpty(bech32)

				var payload map[string]interface{}
				err = json.Unmarshal(ch.Payload, &payload)
				suite.Require().NoError(err)

				data, err = json.Marshal(payload["node_info"])
				suite.Require().NoError(err)

				var expectedNodeInfo cns.NodeInfo
				err = json.Unmarshal(data, &expectedNodeInfo)
				suite.Require().NoError(err)

				suite.Require().Equal(expectedNodeInfo.Bech32Config, bech32)
			}
		})
	}
}
