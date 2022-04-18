package tests

import (
	"github.com/allinbits/emeris-utils/exported/sdktypes"
)

const (
	cachedSupplyEndPoint = "cached/cosmos/bank/v1beta1/supply"
	supplyEndPoint       = "liquidity/cosmos/bank/v1beta1/supply"
	inflationEndPoint    = "liquidity/cosmos/mint/v1beta1/inflation"
)

func (suite testCtx) TestCachedSupply() {
	// get cached supply
	var cachedValues map[string]interface{}
	err := suite.Client.GetJson(&cachedValues, cachedSupplyEndPoint)
	suite.Require().NoError(err)

	// get supply
	var supplyValues map[string]interface{}
	err = suite.Client.GetJson(&supplyValues, supplyEndPoint)
	suite.Require().NoError(err)

	// convert supply to Dec
	supply, err := sdktypes.NewDecFromStr(supplyValues["supply"].([]interface{})[0].(map[string]interface{})["amount"].(string))
	suite.Require().NoError(err)
	cachedSupply, err := sdktypes.NewDecFromStr(cachedValues["supply"].([]interface{})[0].(map[string]interface{})["amount"].(string))
	suite.Require().NoError(err)

	// get mint inflation
	var inflationData map[string]interface{}
	err = suite.Client.GetJson(&inflationData, inflationEndPoint)
	suite.Require().NoError(err)

	inflation, err := sdktypes.NewDecFromStr(inflationData["inflation"].(string))
	suite.Require().NoError(err)

	// supply/cachedSupply should be less than 1.00038
	// considering cosmos-hub inflation as 14% i.e. 0.038% per day
	// for the cached endpoint to be reliable enough that the supply difference isn't more than one day's minted tokens,
	// supply/cachedSupply should not be greater than 1.00038
	supplyToCachedSupplyRation := supply.Quo(cachedSupply)
	threshold := inflation.QuoInt64(365).Add(sdktypes.NewDec(1))
	suite.Require().True(supplyToCachedSupplyRation.LTE(threshold))
}
