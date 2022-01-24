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

const chainFeeTokenEndpoint = "chain/%s/fee/token"

func TestChainFeeToken(t *testing.T) {
	t.Parallel()

	for _, ch := range testCtx.chains {
		t.Run(ch.Name, func(t *testing.T) {
			// arrange
			url := fmt.Sprintf(baseUrl+chainFeeTokenEndpoint, testCtx.emIngress.Protocol, testCtx.emIngress.Host, testCtx.emIngress.APIServerPath, ch.Name)
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

				data, err := json.Marshal(respValues["fee_tokens"])
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

				var expectedFeeDenoms cns.DenomList
				for _, denom := range expectedDenoms {
					if denom.FeeToken {
						expectedFeeDenoms = append(expectedFeeDenoms, denom)
					}
				}
				require.Equal(t, len(expectedFeeDenoms), len(denoms))
				require.ElementsMatch(t, expectedFeeDenoms, denoms)
			}
		})
	}
}
