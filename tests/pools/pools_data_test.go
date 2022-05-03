package tests

import (
	"net/http"

	utils "github.com/allinbits/demeris-backend/test_utils"
)

const (
	poolsEndpoint = "liquidity/cosmos/liquidity/v1beta1/pools"
)

func (suite *testCtx) TestPoolsData() {
	suite.T().Skip("skip: we don't expose cosmos-hub this way anymore (see dexaggregation feature and new LCD endpoints)")

	url := suite.Client.BuildUrl(poolsEndpoint)

	resp, err := suite.Client.Get(url)
	suite.Require().NoError(err)

	suite.Require().Equal(http.StatusOK, resp.StatusCode)

	var respValues map[string]interface{}
	utils.RespBodyToMap(resp.Body, &respValues, suite.T())

	err = resp.Body.Close()
	suite.Require().NoError(err)

	suite.Require().NotNil(respValues)

	pools := respValues["pools"]
	suite.Require().NotEmpty(pools)
}
