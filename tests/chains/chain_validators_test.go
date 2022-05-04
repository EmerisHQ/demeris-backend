package tests

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/allinbits/demeris-backend-models/tracelistener"
	utils "github.com/allinbits/demeris-backend/test_utils"
)

const chainValidatorsEndpoint = "chain/%s/validators"

func (suite *testCtx) TestChainValidators() {
	for _, ch := range suite.Chains {
		suite.Run(ch.ChainName, func() {
			// arrange
			url := suite.Client.BuildUrl(chainValidatorsEndpoint, ch.ChainName)
			// act
			resp, err := suite.Client.Get(url)
			suite.Require().NoError(err)

			// assert
			if !ch.Enabled {
				suite.Require().Equal(http.StatusBadRequest, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.ChainName, resp.StatusCode))
			} else {
				suite.Require().Equal(http.StatusOK, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.ChainName, resp.StatusCode))

				var respValues map[string]interface{}
				utils.RespBodyToMap(resp.Body, &respValues, suite.T())

				err = resp.Body.Close()
				suite.Require().NoError(err)

				suite.Require().NotEmpty(respValues["validators"])

				for _, validator := range respValues["validators"].([]interface{}) {
					var row tracelistener.ValidatorRow
					data, err := json.Marshal(validator)
					suite.Require().NoError(err)

					err = json.Unmarshal(data, &row)
					suite.Require().NoError(err)
				}
			}
		})
	}
}
