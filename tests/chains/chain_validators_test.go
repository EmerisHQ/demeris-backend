package tests

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	chainModels "github.com/allinbits/demeris-api-server/api/chains"
	"github.com/allinbits/demeris-backend-models/tracelistener"
)

const chainValidatorsEndpoint = "chain/%s/validators"

func (suite *testCtx) TestChainValidators() {
	for _, ch := range suite.Chains {
		suite.T().Run(ch.Name, func(t *testing.T) {
			// arrange
			url := suite.Client.BuildUrl(chainValidatorsEndpoint, ch.Name)
			// act
			resp, err := suite.Client.Get(url)
			suite.Require().NoError(err)

			// assert
			if !ch.Enabled {
				suite.Require().Equal(http.StatusBadRequest, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))
			} else {
				suite.Require().Equal(http.StatusOK, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))

				data, err := ioutil.ReadAll(resp.Body)
				suite.Require().NoError(err)

				err = resp.Body.Close()
				suite.Require().NoError(err)

				var val chainModels.ValidatorsResponse
				err = json.Unmarshal(data, &val)
				suite.Require().NoError(err)

				for _, validator := range val.Validators {
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
