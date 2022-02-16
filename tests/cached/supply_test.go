package tests

const (
	cachedSupplyEndPoint = "cached/cosmos/bank/v1beta1/supply"
	supplyEndPoint       = "liquidity/cosmos/bank/v1beta1/supply"
)

func (suite testCtx) TestCachedSupply() {
	suite.T().Skip("FIXME: Skipped until we find a reliable way to verify cached supply (e.g. based on block-height)")
	return

	suite.T().Parallel()

	// get cached supply
	var cachedValues map[string]interface{}
	err := suite.Client.GetJson(&cachedValues, cachedSupplyEndPoint)
	suite.NoError(err)

	// get supply
	var supplyValues map[string]interface{}
	err = suite.Client.GetJson(&supplyValues, supplyEndPoint)
	suite.NoError(err)

	suite.Equal(supplyValues["supply"], cachedValues["supply"])
}
