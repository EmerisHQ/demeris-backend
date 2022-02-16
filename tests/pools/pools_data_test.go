package tests

import (
	"net/http"

	utils "github.com/allinbits/demeris-backend/test_utils"
)

const (
	poolsEndpoint = "liquidity/cosmos/liquidity/v1beta1/pools"
)

func (suite *testCtx) TestPoolsData() {
	suite.T().Parallel()

	url := suite.Client.BuildUrl(poolsEndpoint)

	resp, err := suite.Client.Get(url)
	suite.NoError(err)

	suite.Equal(http.StatusOK, resp.StatusCode)

	var respValues map[string]interface{}
	utils.RespBodyToMap(resp.Body, &respValues, suite.T())

	err = resp.Body.Close()
	suite.NoError(err)

	suite.NotNil(respValues)

	pools, _ := respValues["pools"]
	suite.NotEmpty(pools)
}
