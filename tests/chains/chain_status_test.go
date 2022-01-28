package tests

import (
	"fmt"
	"net/http"
	"testing"

	utils "github.com/allinbits/demeris-backend/test_utils"
)

const (
	statusEndpoint = "chain/%s/status"
	onlineKey      = "online"
)

func (suite *testCtx) TestChainStatus() {
	suite.T().Parallel()

	for _, ch := range suite.chains {
		suite.T().Run(ch.Name, func(t *testing.T) {
			t.Parallel()

			// arrange
			url := fmt.Sprintf(baseUrl+statusEndpoint, suite.emIngress.Protocol, suite.emIngress.Host, suite.emIngress.APIServerPath, ch.Name)
			// act
			resp, err := suite.client.Get(url)
			suite.NoError(err)

			// assert
			if !ch.Enabled {
				suite.Equal(http.StatusBadRequest, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))
			} else {
				suite.Equal(http.StatusOK, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))

				var values map[string]interface{}
				utils.RespBodyToMap(resp.Body, &values, t)

				suite.Equal(true, values[onlineKey].(bool), fmt.Sprintf("Chain %s Online %t", ch.Name, values[onlineKey].(bool)))
			}
		})
	}
}
