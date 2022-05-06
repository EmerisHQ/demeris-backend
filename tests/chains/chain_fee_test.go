package tests

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	chainModels "github.com/emerishq/demeris-api-server/api/chains"
)

const chainFeeEndpoint = "chain/%s/fee"

func (suite *testCtx) TestChainFee() {
	for _, ch := range suite.Chains {
		suite.Run(ch.ChainName, func() {
			// arrange
			url := suite.Client.BuildUrl(chainFeeEndpoint, ch.ChainName)
			// act
			resp, err := suite.Client.Get(url)
			suite.Require().NoError(err)

			// assert
			if !ch.Enabled {
				suite.Require().Equal(http.StatusBadRequest, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.ChainName, resp.StatusCode))
			} else {
				suite.Require().Equal(http.StatusOK, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.ChainName, resp.StatusCode))

				data, err := ioutil.ReadAll(resp.Body)
				suite.Require().NoError(err)

				var denoms chainModels.FeeResponse
				err = json.Unmarshal(data, &denoms)
				suite.Require().NoError(err)
				suite.Require().NotEmpty(denoms)

				suite.Require().Equal(ch.Denoms, denoms.Denoms)
			}
		})
	}
}
