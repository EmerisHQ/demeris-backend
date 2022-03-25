package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/allinbits/demeris-backend-models/cns"
	utils "github.com/allinbits/demeris-backend/test_utils"
)

const chainFeeEndpoint = "chain/%s/fee"

func (suite *testCtx) TestChainFee() {
	for _, ch := range suite.Chains {
		suite.T().Run(ch.Name, func(t *testing.T) {
			// arrange
			url := suite.Client.BuildUrl(chainFeeEndpoint, ch.Name)
			// act
			suite.T().Log(url)
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

				data, err := json.Marshal(respValues["denoms"])
				suite.NoError(err)

				var denoms cns.DenomList
				err = json.Unmarshal(data, &denoms)
				suite.NoError(err)

				suite.NotEmpty(denoms)

				var payload map[string]interface{}
				err = json.Unmarshal(ch.Payload, &payload)
				suite.NoError(err)

				data, err = json.Marshal(payload["denoms"])
				suite.NoError(err)

				var expectedDenoms cns.DenomList
				err = json.Unmarshal(data, &expectedDenoms)
				suite.NoError(err)

				var expectedFeeDenoms cns.DenomList
				for _, denom := range expectedDenoms {
					if denom.FeeToken {
						expectedFeeDenoms = append(expectedFeeDenoms, denom)
					}
				}
				suite.Equal(expectedFeeDenoms, denoms)
			}
		})
	}
}
