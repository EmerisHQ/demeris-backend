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

const (
	primaryChannelsEndpoint = "chain/%s/primary_channels"
	primaryChannelskey      = "primary_channels"
)

func TestPrimaryChannels(t *testing.T) {
	t.Parallel()

	// arrange
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

			defer resp.Body.Close()

			// assert
			if !ch.Enabled {
				require.Equal(t, http.StatusBadRequest, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))
			} else {
				require.Equal(t, http.StatusOK, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))

				var respValues map[string]interface{}
				utils.RespBodyToMap(resp.Body, &respValues, t)

				data, err := json.Marshal(respValues[primaryChannelskey])
				require.NoError(t, err)

				var channels []cns.DbStringMap
				err = json.Unmarshal(data, &channels)
				require.NoError(t, err)

				formattedChannels := make(map[string]string, len(channels))
				for _, channel := range channels {
					formattedChannels[channel["counterparty"]] = channel["channel_name"]
				}

				var payload map[string]interface{}
				err = json.Unmarshal(ch.Payload, &payload)
				require.NoError(t, err)

				data, err = json.Marshal(payload["primary_channel"])
				require.NoError(t, err)

				var expectedChannels map[string]string
				err = json.Unmarshal(data, &expectedChannels)
				require.NoError(t, err)

				require.Equal(t, expectedChannels, formattedChannels)
			}
		})
	}
}
