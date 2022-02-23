package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	utils "github.com/allinbits/demeris-backend/test_utils"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	chainSupplyEndpoint = "chain/%s/supply"
	supplyKey           = "supply"
)

func (suite *testCtx) TestChainSupply() {
	for _, ch := range suite.Chains {
		suite.T().Run(ch.Name, func(t *testing.T) {

			// arrange
			url := suite.Client.BuildUrl(chainSupplyEndpoint, ch.Name)
			// act
			resp, err := suite.Client.Get(url)
			suite.NoError(err)

			defer resp.Body.Close()

			// assert
			if !ch.Enabled {
				suite.Equal(http.StatusBadRequest, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))
			} else {
				suite.Equal(http.StatusOK, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))

				var respValues map[string]interface{}
				utils.RespBodyToMap(resp.Body, &respValues, t)

				data, err := json.Marshal(respValues[supplyKey])
				suite.NoError(err)

				var coins sdk.Coins
				err = json.Unmarshal(data, &coins)
				suite.NoError(err)

				//check if the repsonse is empty
				suite.NotEmpty(coins)
			}
		})
	}
}
