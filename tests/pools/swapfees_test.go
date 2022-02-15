package tests

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/allinbits/demeris-backend-models/api"
	liquiditytypes "github.com/gravity-devs/liquidity/x/liquidity/types"
	tmjson "github.com/tendermint/tendermint/libs/json"
)

const (
	swapfeesendpoint = "pool/%d/swapfees"
)

func (suite *testCtx) TestPoolSwapFees() {
	suite.T().Parallel()

	url := fmt.Sprintf(baseUrl+poolsEndpoint, suite.EmIngress.Protocol, suite.EmIngress.Host, suite.EmIngress.APIServerPath)

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
		url := fmt.Sprintf(baseUrl+swapfeesendpoint, suite.EmIngress.Protocol, suite.EmIngress.Host, suite.EmIngress.APIServerPath, pool.Id)
		resp, err := suite.Client.Get(url)
		suite.Require().NoError(err)
		suite.Require().Equal(http.StatusOK, resp.StatusCode)

		data, err := ioutil.ReadAll(resp.Body)
		suite.Require().NoError(err)

		var fees api.SwapFeesResponse
		suite.Require().NoError(json.Unmarshal(data, &fees))
		suite.Require().NotEmpty(fees.Fees)

		suite.Require().True(fees.Fees.IsAllPositive())

		err = resp.Body.Close()
		suite.Require().NoError(err)
	}
}
