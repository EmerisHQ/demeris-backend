package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/allinbits/demeris-backend-models/tracelistener"
	utils "github.com/allinbits/demeris-backend/test_utils"
)

const chainValidatorsEndpoint = "chain/%s/validators"

func (suite *testCtx) TestChainValidators() {
	env := os.Getenv("ENV")
	if strings.ToLower(env) == "dev" {
		suite.T().Skip("FIXME: Skipping in DEV. Enable after recreating the environment")
		return
	}

	suite.T().Parallel()

	for _, ch := range suite.Chains {
		suite.T().Run(ch.Name, func(t *testing.T) {
			// arrange
			url := fmt.Sprintf(baseUrl+chainValidatorsEndpoint, suite.EmIngress.Protocol, suite.EmIngress.Host, suite.EmIngress.APIServerPath, ch.Name)
			// act
			resp, err := suite.Client.Get(url)
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
