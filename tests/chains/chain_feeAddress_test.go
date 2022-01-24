package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	utils "github.com/allinbits/demeris-backend/test_utils"
)

const chainFeeAddressEndpoint = "chain/%s/fee/address"

func TestChainFeeAddress(t *testing.T) {
	t.Parallel()

	for _, ch := range testCtx.chains {
		t.Run(ch.Name, func(t *testing.T) {
			// arrange
			url := fmt.Sprintf(baseUrl+chainFeeAddressEndpoint, testCtx.emIngress.Protocol, testCtx.emIngress.Host, testCtx.emIngress.APIServerPath, ch.Name)
			// act
			resp, err := testCtx.client.Get(url)
			require.NoError(t, err)

			// assert
			if !ch.Enabled {
				require.Equal(t, http.StatusBadRequest, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))
			} else {
				require.Equal(t, http.StatusOK, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))

				var payload map[string]interface{}
				err := json.Unmarshal(ch.Payload, &payload)
				require.NoError(t, err)

				var respValues map[string]interface{}
				utils.RespBodyToMap(resp.Body, &respValues, t)

				err = resp.Body.Close()
				require.NoError(t, err)

				require.Equal(t, payload["demeris_addresses"], respValues["fee_address"])
			}
		})
	}
}
