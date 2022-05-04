package tests

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	api "github.com/emerishq/demeris-api-server/api/liquidity"
	liquiditytypes "github.com/gravity-devs/liquidity/x/liquidity/types"
	tmjson "github.com/tendermint/tendermint/libs/json"
)

const (
	swapfeesendpoint = "pool/%d/swapfees"
)

func (suite *testCtx) TestPoolSwapFees() {
	url := suite.Client.BuildUrl(poolsEndpoint)

	resp, err := suite.Client.Get(url)
	suite.Require().NoError(err)

	suite.Require().Equal(http.StatusOK, resp.StatusCode)

	data, err := ioutil.ReadAll(resp.Body)
	suite.Require().NoError(err)

	var poolsRes liquiditytypes.QueryLiquidityPoolsResponse
	suite.Require().NoError(tmjson.Unmarshal(data, &poolsRes))
	suite.Require().NotEmpty(poolsRes.Pools)

	err = resp.Body.Close()
	suite.Require().NoError(err)

	for _, pool := range poolsRes.Pools {
		url := suite.Client.BuildUrl(swapfeesendpoint, pool.Id)
		resp, err := suite.Client.Get(url)
		suite.Require().NoError(err)
		suite.Require().Equal(http.StatusOK, resp.StatusCode)

		data, err := ioutil.ReadAll(resp.Body)
		suite.Require().NoError(err)

		var fees api.SwapFeesResponse
		suite.Require().NoError(json.Unmarshal(data, &fees))
		suite.Require().NotEmpty(fees)
		if len(fees.Fees) != 0 {
			suite.NotEmpty(fees.Fees, fmt.Sprintf("Pool:%d", pool.Id))
			suite.True(fees.Fees.IsAllPositive(), fmt.Sprintf("Pool:%d", pool.Id))
		}

		err = resp.Body.Close()
		suite.Require().NoError(err)
	}
}
