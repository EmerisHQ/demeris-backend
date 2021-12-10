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
	primaryChannelsEndpoint = "chain/%s/primary_channels"
	primaryChannelskey      = "primary_channels"
)

func TestPrimaryChannels(t *testing.T) {
	t.Parallel()

	// arrange
	os.Setenv("ENV", "dev")
	env := os.Getenv("ENV")
	emIngress, _ := utils.LoadIngressInfo(env, t)
	require.NotNil(t, emIngress)

	chains := utils.LoadChainsInfo(env, t)
	require.NotNil(t, chains)

	client := utils.CreateNetClient(env, t)
	require.NotNil(t, client)

	for _, ch := range chains {
		t.Run(ch.Name, func(t *testing.T) {

			// arrange
			url := fmt.Sprintf(baseUrl+primaryChannelsEndpoint, emIngress.Protocol, emIngress.Host, emIngress.APIServerPath, ch.Name)
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

				//expect a non empty data
				channels := respValues[primaryChannelskey]
				require.NotEmpty(t, channels)
			}
		})
	}
}
