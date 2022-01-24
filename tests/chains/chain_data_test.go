package tests

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	utils "github.com/allinbits/demeris-backend/test_utils"
)

const (
	chainEndpoint = "chain/%s"
	respChainKey  = "chain"
)

func TestChainData(t *testing.T) {
	t.Parallel()

	for _, ch := range testCtx.chains {
		t.Run(ch.Name, func(t *testing.T) {
			//t.Parallel()

			// arrange
			url := fmt.Sprintf(baseUrl+chainEndpoint, testCtx.emIngress.Protocol, testCtx.emIngress.Host, testCtx.emIngress.APIServerPath, ch.Name)
			// act
			resp, err := testCtx.client.Get(url)
			require.NoError(t, err)

			// assert
			if !ch.Enabled {
				require.Equal(t, http.StatusBadRequest, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))
			} else {
				require.Equal(t, http.StatusOK, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))

				var respValues map[string]interface{}
				var expValues map[string]interface{}
				utils.RespBodyToMap(resp.Body, &respValues, t)
				utils.StringToMap(ch.Payload, &expValues, t)

				// response is nested one level down
				require.Equal(t, expValues, respValues[respChainKey].(map[string]interface{}))
			}
		})
	}
}
