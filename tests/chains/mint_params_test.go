package tests

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	utils "github.com/allinbits/demeris-backend/test_utils"
)

const (
	mintParamsEndpoint = "chain/%s/mint/params"
	paramsKey          = "params"
)

func TestMintParams(t *testing.T) {
	t.Parallel()

	for _, ch := range testCtx.chains {
		t.Run(ch.Name, func(t *testing.T) {

			// arrange
			url := fmt.Sprintf(baseUrl+mintParamsEndpoint, testCtx.emIngress.Protocol, testCtx.emIngress.Host, testCtx.emIngress.APIServerPath, ch.Name)
			// act
			resp, err := testCtx.client.Get(url)
			require.NoError(t, err)

			defer resp.Body.Close()

			// assert
			if !ch.Enabled {
				require.Equal(t, http.StatusBadRequest, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))
			} else {
				require.Equal(t, http.StatusOK, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))

				var respValues map[string]interface{}
				utils.RespBodyToMap(resp.Body, &respValues, t)

				//expect a non empty data
				params := respValues[paramsKey]
				require.NotEmpty(t, params)
			}
		})
	}
}
