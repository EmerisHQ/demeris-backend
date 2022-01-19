package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	utils "github.com/allinbits/demeris-backend/test_utils"
)

const chainsEndpoint = "chains"

func TestChainsData(t *testing.T) {
	t.Parallel()

	// arrange
	env := os.Getenv("ENV")
	emIngress, _ := utils.LoadIngressInfo(env, t)
	require.NotNil(t, emIngress.Host)
	chains := utils.LoadChainsInfo(env, t)
	require.NotEmpty(t, chains)
	client := utils.CreateNetClient(env, t)

	// arrange
	url := fmt.Sprintf(baseUrl+chainsEndpoint, emIngress.Protocol, emIngress.Host, emIngress.APIServerPath)
	// act
	resp, err := client.Get(url)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var respValues map[string]interface{}
	utils.RespBodyToMap(resp.Body, &respValues, t)

	err = resp.Body.Close()
	require.NoError(t, err)

	expValues := make(map[string][]map[string]interface{}, 0)
	for _, ch := range chains {
		if ch.Enabled {
			chainUrl := fmt.Sprintf(baseUrl+"chain/%s", emIngress.Protocol, emIngress.Host, emIngress.APIServerPath, ch.Name)
			chainResp, err := client.Get(chainUrl)
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