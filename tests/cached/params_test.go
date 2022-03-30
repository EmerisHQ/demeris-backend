package tests

import (
	liquiditytypes "github.com/gravity-devs/liquidity/x/liquidity/types"
)

const (
	cachedParamsEndPoint    = "cached/cosmos/liquidity/v1beta1/params"
	liquidityParamsEndPoint = "liquidity/cosmos/liquidity/v1beta1/params"
)

func (suite *testCtx) TestCachedParams() {
	suite.T().Parallel()

	// get cached params
	var cachedValues liquiditytypes.QueryParamsResponse
	err := suite.Client.GetJson(&cachedValues, cachedParamsEndPoint)
	suite.Require().NoError(err)

	// get liquidity params
	var liquidityValues liquiditytypes.QueryParamsResponse
	err = suite.Client.GetJson(&liquidityValues, liquidityParamsEndPoint)
	suite.Require().NoError(err)

	suite.Require().Equal(liquidityValues.Params, cachedValues.Params)
}
