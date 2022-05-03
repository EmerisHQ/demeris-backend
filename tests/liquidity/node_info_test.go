package tests

import (
	"fmt"
	"net/http"

	utils "github.com/allinbits/demeris-backend/test_utils"
)

const (
	liquidityNodeEndpoint = "/liquidity/node_info"
	chainName             = "cosmos-hub"
)

func (suite *testCtx) TestLiquidityStatus() {
	suite.T().Skip("skip: this test is comparing ch.NodeInfo against its own field ch.NodeInfo.ChainID, not sure what the test should be doing instead")

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

				// v, ok := values["node_info"].(map[string]interface{})
				// suite.True(ok)
				networkName := ch.NodeInfo

				// var fileResp map[string]interface{}
				// utils.StringToMap(ch.Payload, &fileResp, t)

				// fv, ok := fileResp["node_info"].(map[string]interface{})
				// suite.True(ok)

				expectedName := ch.NodeInfo.ChainID

				suite.Require().Equal(expectedName, networkName)
			})
		}
	}
}
