package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/allinbits/demeris-backend-models/cns"
	utils "github.com/allinbits/demeris-backend/test_utils"
)

const chainBech32Endpoint = "chain/%s/bech32"

func TestChainBech32(t *testing.T) {
	t.Parallel()

	for _, ch := range testCtx.chains {
		t.Run(ch.Name, func(t *testing.T) {
			// arrange
			url := fmt.Sprintf(baseUrl+chainBech32Endpoint, testCtx.emIngress.Protocol, testCtx.emIngress.Host, testCtx.emIngress.APIServerPath, ch.Name)
			// act
			resp, err := testCtx.client.Get(url)
			require.NoError(t, err)

			// assert
			if !ch.Enabled {
				require.Equal(t, http.StatusBadRequest, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))
			} else {
				require.Equal(t, http.StatusOK, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))

				var respValues map[string]interface{}
				utils.RespBodyToMap(resp.Body, &respValues, t)

				err = resp.Body.Close()
				require.NoError(t, err)

				data, err := json.Marshal(respValues["bech32_config"])
				require.NoError(t, err)

				var bech32 cns.Bech32Config
				err = json.Unmarshal(data, &bech32)
				require.NoError(t, err)

				require.NotEmpty(t, bech32)

				var payload map[string]interface{}
				err = json.Unmarshal(ch.Payload, &payload)
				require.NoError(t, err)

				data, err = json.Marshal(payload["node_info"])
				require.NoError(t, err)

				var expectedNodeInfo cns.NodeInfo
				err = json.Unmarshal(data, &expectedNodeInfo)
				require.NoError(t, err)

				require.Equal(t, expectedNodeInfo.Bech32Config, bech32)
			}
		})
	}
}
