package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/allinbits/demeris-backend-models/tracelistener"
	utils "github.com/allinbits/demeris-backend/test_utils"
)

const chainValidatorsEndpoint = "chain/%s/validators"

func TestChainValidators(t *testing.T) {
	t.Parallel()

	for _, ch := range testCtx.chains {
		t.Run(ch.Name, func(t *testing.T) {
			// arrange
			url := fmt.Sprintf(baseUrl+chainValidatorsEndpoint, testCtx.emIngress.Protocol, testCtx.emIngress.Host, testCtx.emIngress.APIServerPath, ch.Name)
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

				require.NotEmpty(t, respValues["validators"])

				for _, validator := range respValues["validators"].([]interface{}) {
					var row tracelistener.ValidatorRow
					data, err := json.Marshal(validator)
					require.NoError(t, err)

					err = json.Unmarshal(data, &row)
					require.NoError(t, err)
				}
			}
		})
	}
}
