package tests

import (
	"fmt"
	"net/http"

	utils "github.com/emerishq/demeris-backend/test_utils"
)

const (
	liquidityNodeEndpoint = "/liquidity/node_info"
	chainName             = "cosmos-hub"
)

func (suite *testCtx) TestLiquidityStatus() {

	for _, ch := range suite.Chains {
		if ch.ChainName == chainName {
			suite.Run(ch.ChainName, func() {

				// arrange
				url := suite.Client.BuildUrl(liquidityNodeEndpoint)
				// act
				resp, err := suite.Client.Get(url)
				suite.Require().NoError(err)

				defer resp.Body.Close()

				// assert
				suite.Require().Equal(http.StatusOK, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.ChainName, resp.StatusCode))

				var values map[string]interface{}
				utils.RespBodyToMap(resp.Body, &values, suite.T())
				v, ok := values["node_info"].(map[string]interface{})
				suite.Require().True(ok)

				networkName := v["network"]

				expectedName := ch.NodeInfo.ChainID
				suite.Require().Equal(expectedName, networkName)
			})
		}
	}
}
