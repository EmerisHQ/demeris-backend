package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/allinbits/demeris-backend-models/cns"
	utils "github.com/allinbits/demeris-backend/test_utils"
)

const chainFeeTokenEndpoint = "chain/%s/fee/token"

func (suite *testCtx) TestChainFeeToken() {
	for _, ch := range suite.Chains {
		suite.T().Run(ch.Name, func(t *testing.T) {
			// arrange
			url := suite.Client.BuildUrl(chainFeeTokenEndpoint, ch.Name)
			// act
			resp, err := suite.Client.Get(url)
			suite.Require().NoError(err)

			// assert
			if !ch.Enabled {
				suite.Require().Equal(http.StatusBadRequest, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))
			} else {
				suite.Require().Equal(http.StatusOK, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))

				var respValues map[string]interface{}
				utils.RespBodyToMap(resp.Body, &respValues, t)

				err = resp.Body.Close()
				suite.Require().NoError(err)

				data, err := json.Marshal(respValues["fee_tokens"])
				suite.Require().NoError(err)

				var denoms cns.DenomList
				err = json.Unmarshal(data, &denoms)
				suite.Require().NoError(err)

				suite.Require().NotEmpty(denoms)

				var payload map[string]interface{}
				err = json.Unmarshal(ch.Payload, &payload)
				suite.Require().NoError(err)

				data, err = json.Marshal(payload["denoms"])
				suite.Require().NoError(err)

				var expectedDenoms cns.DenomList
				err = json.Unmarshal(data, &expectedDenoms)
				suite.Require().NoError(err)

				var expectedFeeDenoms cns.DenomList
				for _, denom := range expectedDenoms {
					if denom.FeeToken {
						expectedFeeDenoms = append(expectedFeeDenoms, denom)
					}
				}

				suite.Require().Equal(len(expectedFeeDenoms), len(denoms))
				suite.Require().EqualValues(expectedFeeDenoms, denoms)
			}
		})
	}
}
