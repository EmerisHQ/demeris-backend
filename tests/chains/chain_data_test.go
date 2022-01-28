package tests

import (
	"fmt"
	"net/http"
	"testing"

	utils "github.com/allinbits/demeris-backend/test_utils"
)

const (
	chainEndpoint = "chain/%s"
	respChainKey  = "chain"
)

func (suite *testCtx) TestChainData() {
	suite.T().Parallel()

	for _, ch := range suite.chains {
		suite.T().Run(ch.Name, func(t *testing.T) {
			//t.Parallel()

			// arrange
			url := fmt.Sprintf(baseUrl+chainEndpoint, suite.emIngress.Protocol, suite.emIngress.Host, suite.emIngress.APIServerPath, ch.Name)
			// act
			resp, err := suite.client.Get(url)
			suite.NoError(err)

			// assert
			if !ch.Enabled {
				suite.Equal(http.StatusBadRequest, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))
			} else {
				suite.Equal(http.StatusOK, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))

				var respValues map[string]interface{}
				var expValues map[string]interface{}
				utils.RespBodyToMap(resp.Body, &respValues, t)
				utils.StringToMap(ch.Payload, &expValues, t)

				// response is nested one level down
				suite.Equal(expValues, respValues[respChainKey].(map[string]interface{}))
			}
		})
	}
}
