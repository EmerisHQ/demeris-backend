package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	utils "github.com/allinbits/demeris-backend/test_utils"
	"github.com/emerishq/demeris-backend-models/cns"
)

const chainFeeTokenEndpoint = "chain/%s/fee/token"

func (suite *testCtx) TestChainFeeToken() {
	for _, ch := range suite.Chains {
		suite.T().Run(ch.ChainName, func(t *testing.T) {
			// arrange
			url := suite.Client.BuildUrl(chainFeeTokenEndpoint, ch.ChainName)
			// act
			resp, err := suite.Client.Get(url)
			suite.Require().NoError(err)

			// assert
			if !ch.Enabled {
				suite.Require().Equal(http.StatusBadRequest, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.ChainName, resp.StatusCode))
			} else {
				suite.Require().Equal(http.StatusOK, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.ChainName, resp.StatusCode))

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

				expectedDenoms := ch.Denoms
				var expectedFeeDenoms cns.DenomList
				for _, denom := range expectedDenoms {
					if denom.FeeToken {
						expectedFeeDenoms = append(expectedFeeDenoms, denom)
					}
				}

				suite.Require().Equal(len(expectedFeeDenoms), len(denoms))
				suite.Require().Equal(expectedFeeDenoms, denoms)
			}
		})
	}
}
