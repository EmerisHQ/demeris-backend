package tests

import (
	"fmt"
	"strings"

	utils "github.com/allinbits/demeris-backend/test_utils"
)

const (
	cachedSupplyEndPoint = "cached/cosmos/bank/v1beta1/supply"
	supplyEndPoint       = "liquidity/cosmos/bank/v1beta1/supply"
)

func (suite testCtx) TestCachedSupply() {
	suite.T().Skip("FIXME: Skipped until we find a reliable way to verify cached supply (e.g. based on block-height)")
	return

	suite.T().Parallel()

	// get cached supply
	urlPattern := strings.Join([]string{baseUrl, cachedSupplyEndPoint}, "")

	url := fmt.Sprintf(urlPattern, suite.emIngress.Protocol, suite.emIngress.Host, suite.emIngress.APIServerPath)
	cachedResp, err := suite.client.Get(url)
	suite.NoError(err)

	defer cachedResp.Body.Close()

	var cachedValues map[string]interface{}
	utils.RespBodyToMap(cachedResp.Body, &cachedValues, suite.T())

	// get supply
	urlPattern = strings.Join([]string{baseUrl, supplyEndPoint}, "")
	url = fmt.Sprintf(urlPattern, suite.emIngress.Protocol, suite.emIngress.Host, suite.emIngress.APIServerPath)
	supplyResp, err := suite.client.Get(url)
	suite.NoError(err)

	defer supplyResp.Body.Close()

	var supplyValues map[string]interface{}
	utils.RespBodyToMap(supplyResp.Body, &supplyValues, suite.T())

	suite.Equal(supplyValues["supply"], cachedValues["supply"])
}
