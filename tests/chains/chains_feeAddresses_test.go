package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	utils "github.com/allinbits/demeris-backend/test_utils"
)

const chainsFeeAddressesEndpoint = "chains/fee/addresses"

func TestChainsFeeAddresses(t *testing.T) {
	t.Parallel()

	// arrange
	url := fmt.Sprintf(baseUrl+chainsFeeAddressesEndpoint, testCtx.emIngress.Protocol, testCtx.emIngress.Host, testCtx.emIngress.APIServerPath)
	// act
	resp, err := testCtx.client.Get(url)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var respValues map[string]interface{}
	utils.RespBodyToMap(resp.Body, &respValues, t)

	err = resp.Body.Close()
	require.NoError(t, err)

	expValues := make(map[string][]map[string]interface{}, 0)
	for _, ch := range testCtx.chains {
		if ch.Enabled {
			var payload map[string]interface{}
			err := json.Unmarshal(ch.Payload, &payload)
			require.NoError(t, err)

			expValues["fee_addresses"] = append(expValues["fee_addresses"], map[string]interface{}{
				"chain_name":  ch.Name,
				"fee_address": payload["demeris_addresses"],
			})
		}
	}

	expValuesData, err := json.Marshal(expValues)
	require.NoError(t, err)

	var expValuesInterface map[string]interface{}
	err = json.Unmarshal(expValuesData, &expValuesInterface)
	require.NoError(t, err)

	require.ElementsMatch(t, expValuesInterface["fee_addresses"], respValues["fee_addresses"])
}
