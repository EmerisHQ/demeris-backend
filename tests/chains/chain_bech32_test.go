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
			suite.T().Log(url)
			resp, err := suite.Client.Get(url)
			suite.NoError(err)

			// assert
			if !ch.Enabled {
				suite.Equal(http.StatusBadRequest, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))
			} else {
				suite.Equal(http.StatusOK, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))

				var respValues map[string]interface{}
				utils.RespBodyToMap(resp.Body, &respValues, t)

				err = resp.Body.Close()
				suite.NoError(err)

				data, err := json.Marshal(respValues["bech32_config"])
				suite.NoError(err)

				var bech32 cns.Bech32Config
				err = json.Unmarshal(data, &bech32)
				suite.NoError(err)

				suite.NotEmpty(bech32)

				var payload map[string]interface{}
				err = json.Unmarshal(ch.Payload, &payload)
				suite.NoError(err)

				data, err = json.Marshal(payload["node_info"])
				suite.NoError(err)

				var expectedNodeInfo cns.NodeInfo
				err = json.Unmarshal(data, &expectedNodeInfo)
				suite.NoError(err)

				suite.Equal(expectedNodeInfo.Bech32Config, bech32)
			}
		})
	}
}
