package tests

import (
	"fmt"
	"net/http"

	utils "github.com/allinbits/demeris-backend/test_utils"
)

const (
	mintParamsEndpoint = "chain/%s/mint/params"
	paramsKey          = "params"
)

func (suite *testCtx) TestMintParams() {
	for _, ch := range suite.Chains {
		suite.Run(ch.ChainName, func() {
			if ch.ChainName == "crypto-org" {
				suite.T().Skip("skip: crypto-org, api-server returns error")
			}

			// arrange
			url := suite.Client.BuildUrl(mintParamsEndpoint, ch.ChainName)
			// act
			resp, err := suite.Client.Get(url)
			suite.Require().NoError(err)

			defer resp.Body.Close()

			// assert
			if !ch.Enabled {
				suite.Require().Equal(http.StatusBadRequest, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.ChainName, resp.StatusCode))
			} else {
				suite.Require().Equal(http.StatusOK, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.ChainName, resp.StatusCode))

				var respValues map[string]interface{}
				utils.RespBodyToMap(resp.Body, &respValues, suite.T())

				//expect a non empty data
				params := respValues[paramsKey]
				suite.Require().NotEmpty(params)
			}
		})
	}
}
