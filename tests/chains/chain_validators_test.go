package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/allinbits/demeris-backend-models/tracelistener"
	utils "github.com/allinbits/demeris-backend/test_utils"
)

const chainValidatorsEndpoint = "chain/%s/validators"

func (suite *testCtx) TestChainValidators() {
	suite.T().Parallel()

	for _, ch := range suite.chains {
		suite.T().Run(ch.Name, func(t *testing.T) {
			// arrange
			url := fmt.Sprintf(baseUrl+chainValidatorsEndpoint, suite.emIngress.Protocol, suite.emIngress.Host, suite.emIngress.APIServerPath, ch.Name)
			// act
			resp, err := suite.client.Get(url)
			suite.NoError(err)

			// assert
			if !ch.Enabled {
				suite.Equal(http.StatusBadRequest, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))
			} else {
				suite.Equal(http.StatusOK, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))

				var respValues map[string]interface{}
				utils.RespBodyToMap(resp.Body, &respValues, t)

				err = resp.Body.Close()
				suite.NoError(err)

				suite.NotEmpty(respValues["validators"])

				for _, validator := range respValues["validators"].([]interface{}) {
					var row tracelistener.ValidatorRow
					data, err := json.Marshal(validator)
					suite.NoError(err)

					err = json.Unmarshal(data, &row)
					suite.NoError(err)
				}
			}
		})
	}
}
