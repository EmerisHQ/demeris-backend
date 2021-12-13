package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/allinbits/demeris-backend-models/cns"
	utils "github.com/allinbits/demeris-backend/test_utils"
)

const chainFeeEndpoint = "chain/%s/fee"

func TestChainFee(t *testing.T) {
	t.Parallel()

	// arrange
	env := os.Getenv("ENV")
	emIngress, _ := utils.LoadIngressInfo(env, t)
	chains := utils.LoadChainsInfo(env, t)
	client := utils.CreateNetClient(env, t)

	for _, ch := range chains {
		t.Run(ch.Name, func(t *testing.T) {
			// arrange
			url := fmt.Sprintf(baseUrl+chainFeeEndpoint, emIngress.Protocol, emIngress.Host, emIngress.APIServerPath, ch.Name)
			// act
			resp, err := client.Get(url)
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

				data, err := json.Marshal(respValues["denoms"])
				require.NoError(t, err)

				var denoms cns.DenomList
				err = json.Unmarshal(data, &denoms)
				require.NoError(t, err)

				require.NotEmpty(t, denoms)

				var payload map[string]interface{}
				err = json.Unmarshal(ch.Payload, &payload)
				require.NoError(t, err)

				data, err = json.Marshal(payload["denoms"])
				require.NoError(t, err)

				var expectedDenoms cns.DenomList
				err = json.Unmarshal(data, &expectedDenoms)
				require.NoError(t, err)

				require.ElementsMatch(t, expectedDenoms, denoms)
			}
		})
	}
}
