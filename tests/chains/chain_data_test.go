package tests

import (
	"fmt"
	"net/http"
	"os"
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

	// arrange
	env := os.Getenv("ENV")
	emIngress, _ := utils.LoadIngressInfo(env, t)
	chains := utils.LoadChainsInfo(env, t)
	client := utils.CreateNetClient(env, t)

	for _, ch := range chains {
		t.Run(ch.Name, func(t *testing.T) {
			//t.Parallel()

			// arrange
			url := fmt.Sprintf(baseUrl+chainEndpoint, emIngress.Protocol, emIngress.Host, emIngress.APIServerPath, ch.Name)
			// act
			resp, err := client.Get(url)
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
