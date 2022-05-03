package tests

import (
	"io/ioutil"

	liquiditytypes "github.com/gravity-devs/liquidity/x/liquidity/types"
	tmjson "github.com/tendermint/tendermint/libs/json"
)

const (
	cachedPoolsEndPoint    = "cached/cosmos/liquidity/v1beta1/pools"
	liquidityPoolsEndPoint = "liquidity/cosmos/liquidity/v1beta1/pools"
)

func (suite *testCtx) TestCachedPools() {
	suite.T().Skip("skip: we don't expose cosmos-hub this way anymore (see dexaggregation feature and new LCD endpoints)")

	// get cached pools
	url := suite.Client.BuildUrl(cachedPoolsEndPoint)
	cachedResp, err := suite.Client.Get(url)
	suite.Require().NoError(err)

	defer cachedResp.Body.Close()

	var cachedValues liquiditytypes.QueryLiquidityPoolsResponse
	body, err := ioutil.ReadAll(cachedResp.Body)
	suite.Require().NoError(err)

	suite.NoError(tmjson.Unmarshal(body, &cachedValues))

	// get liquidity pools
	liquidityUrl := suite.Client.BuildUrl(liquidityPoolsEndPoint)
	liquidityResp, err := suite.Client.Get(liquidityUrl)
	suite.Require().NoError(err)

	defer liquidityResp.Body.Close()

	var liquidityValues liquiditytypes.QueryLiquidityPoolsResponse
	body, err = ioutil.ReadAll(liquidityResp.Body)
	suite.Require().NoError(err)

	suite.Require().NoError(tmjson.Unmarshal(body, &liquidityValues))

	suite.Require().Equal(liquidityValues.Pools, cachedValues.Pools)
}
