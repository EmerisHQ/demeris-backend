package tests

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	liquiditytypes "github.com/gravity-devs/liquidity/x/liquidity/types"
)

const (
	cachedParamsEndPoint    = "cached/cosmos/liquidity/v1beta1/params"
	liquidityParamsEndPoint = "liquidity/cosmos/liquidity/v1beta1/params"
)

func (suite *testCtx) TestCachedParams() {
	suite.T().Parallel()

	// get cached params
	urlPattern := strings.Join([]string{baseUrl, cachedParamsEndPoint}, "")

	url := fmt.Sprintf(urlPattern, suite.EmIngress.Protocol, suite.EmIngress.Host, suite.EmIngress.APIServerPath)
	cachedResp, err := suite.Client.Get(url)
	suite.NoError(err)

	defer cachedResp.Body.Close()

	var cachedValues liquiditytypes.QueryParamsResponse
	body, err := ioutil.ReadAll(cachedResp.Body)
	suite.NoError(err)

	err = json.Unmarshal(body, &cachedValues)
	suite.NoError(err)

	// get liquidity params
	urlPattern = strings.Join([]string{baseUrl, liquidityParamsEndPoint}, "")

	url = fmt.Sprintf(urlPattern, suite.EmIngress.Protocol, suite.EmIngress.Host, suite.EmIngress.APIServerPath)
	liquidityResp, err := suite.Client.Get(url)
	suite.NoError(err)

	defer liquidityResp.Body.Close()

	var liquidityValues liquiditytypes.QueryParamsResponse
	body, err = ioutil.ReadAll(liquidityResp.Body)
	suite.NoError(err)

	err = json.Unmarshal(body, &liquidityValues)
	suite.NoError(err)

	suite.Equal(liquidityValues.Params, cachedValues.Params)
}
