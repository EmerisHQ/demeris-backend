package tests

import (
	"fmt"
	"io/ioutil"
	"strings"

	liquiditytypes "github.com/gravity-devs/liquidity/x/liquidity/types"
	tmjson "github.com/tendermint/tendermint/libs/json"
)

const (
	cachedPoolsEndPoint    = "cached/cosmos/liquidity/v1beta1/pools"
	liquidityPoolsEndPoint = "liquidity/cosmos/liquidity/v1beta1/pools"
)

func (suite *testCtx) TestCachedPools() {
	suite.T().Parallel()

	// get cached pools
	urlPattern := strings.Join([]string{baseUrl, cachedPoolsEndPoint}, "")
	url := fmt.Sprintf(urlPattern, suite.EmIngress.Protocol, suite.EmIngress.Host, suite.EmIngress.APIServerPath)
	cachedResp, err := suite.Client.Get(url)
	suite.NoError(err)

	defer cachedResp.Body.Close()

	var cachedValues liquiditytypes.QueryLiquidityPoolsResponse
	body, err := ioutil.ReadAll(cachedResp.Body)
	suite.NoError(err)

	suite.NoError(tmjson.Unmarshal(body, &cachedValues))

	// get liquidity pools
	urlPattern = strings.Join([]string{baseUrl, liquidityPoolsEndPoint}, "")

	url = fmt.Sprintf(urlPattern, suite.EmIngress.Protocol, suite.EmIngress.Host, suite.EmIngress.APIServerPath)
	liquidityResp, err := suite.Client.Get(url)
	suite.NoError(err)

	defer liquidityResp.Body.Close()

	var liquidityValues liquiditytypes.QueryLiquidityPoolsResponse
	body, err = ioutil.ReadAll(liquidityResp.Body)
	suite.NoError(err)

	suite.NoError(tmjson.Unmarshal(body, &liquidityValues))

	suite.Equal(liquidityValues.Pools, cachedValues.Pools)
}
