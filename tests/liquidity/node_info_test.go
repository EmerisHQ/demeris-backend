package tests

import (
	"fmt"
	"net/http"
	"testing"

	utils "github.com/allinbits/demeris-backend/test_utils"
)

const (
	liquidityNodeEndpoint = "/liquidity/node_info"
	chainName             = "cosmos-hub"
)

func (suite *testCtx) TestLiquidityStatus() {
	suite.T().Parallel()

	for _, ch := range suite.Chains {
		if ch.Name == chainName {
			suite.T().Run(ch.Name, func(t *testing.T) {

				// arrange
				url := fmt.Sprintf(baseUrl+liquidityNodeEndpoint, suite.EmIngress.Protocol, suite.EmIngress.Host, suite.EmIngress.APIServerPath)
				// act
				resp, err := suite.Client.Get(url)
				suite.NoError(err)

				defer resp.Body.Close()

				// assert
				suite.Equal(http.StatusOK, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))

				var values map[string]interface{}
				utils.RespBodyToMap(resp.Body, &values, t)

				v, ok := values["node_info"].(map[string]interface{})
				suite.True(ok)
				networkName := v["network"]

				var fileResp map[string]interface{}
				utils.StringToMap(ch.Payload, &fileResp, t)

				fv, ok := fileResp["node_info"].(map[string]interface{})
				suite.True(ok)

				expectedName := fv["chain_id"]

				suite.Equal(expectedName, networkName)
			})
		}
	}
}
