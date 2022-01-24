package tests

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	utils "github.com/allinbits/demeris-backend/test_utils"
)

const (
	liquidityNodeEndpoint = "/liquidity/node_info"
	chainName             = "cosmos-hub"
)

func TestLiquidityStatus(t *testing.T) {
	t.Parallel()

	for _, ch := range testCtx.chains {
		if ch.Name == chainName {
			t.Run(ch.Name, func(t *testing.T) {

				// arrange
				url := fmt.Sprintf(baseUrl+liquidityNodeEndpoint, testCtx.emIngress.Protocol, testCtx.emIngress.Host, testCtx.emIngress.APIServerPath)
				// act
				resp, err := testCtx.client.Get(url)
				require.NoError(t, err)

				defer resp.Body.Close()

				// assert
				require.Equal(t, http.StatusOK, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))

				var values map[string]interface{}
				utils.RespBodyToMap(resp.Body, &values, t)

				v, ok := values["node_info"].(map[string]interface{})
				require.True(t, ok)
				networkName := v["network"]

				var fileResp map[string]interface{}
				utils.StringToMap(ch.Payload, &fileResp, t)

				fv, ok := fileResp["node_info"].(map[string]interface{})
				require.True(t, ok)

				expectedName := fv["chain_id"]

				require.Equal(t, expectedName, networkName)
			})
		}
	}
}
