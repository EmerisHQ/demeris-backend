package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	utils "github.com/allinbits/demeris-backend/test_utils"
)

const chainsEndpoint = "chains"

func TestChainsData(t *testing.T) {
	t.Parallel()

	// arrange
	url := fmt.Sprintf(baseUrl+chainsEndpoint, testCtx.emIngress.Protocol, testCtx.emIngress.Host, testCtx.emIngress.APIServerPath)
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
			chainUrl := fmt.Sprintf(baseUrl+"chain/%s", testCtx.emIngress.Protocol, testCtx.emIngress.Host, testCtx.emIngress.APIServerPath, ch.Name)
			chainResp, err := testCtx.client.Get(chainUrl)
			require.NoError(t, err)

			require.Equal(t, http.StatusOK, chainResp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, chainResp.StatusCode))

			var payload map[string]interface{}
			err = json.Unmarshal(ch.Payload, &payload)
			require.NoError(t, err)

			expValues["chains"] = append(expValues["chains"], map[string]interface{}{
				"chain_name":   ch.Name,
				"display_name": payload["display_name"],
				"logo":         payload["logo"],
			})
		}
	}

	expValuesData, err := json.Marshal(expValues)
	require.NoError(t, err)

	var expValuesInterface map[string]interface{}
	err = json.Unmarshal(expValuesData, &expValuesInterface)
	require.NoError(t, err)

	require.ElementsMatch(t, expValuesInterface["chains"], respValues["chains"])
}
