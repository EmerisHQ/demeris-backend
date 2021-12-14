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

const (
	primaryChannelEndpoint = "/chain/%s/primary_channels"
	primaryChannelskey     = "primary_channels"
	channelCounterparty    = "/chain/%s/primary_channel/%s"
)

type PrimaryChannels struct {
	PrimaryChannels []struct {
		Counterparty string `json:"counterparty"`
		ChannelName  string `json:"channel_name"`
	} `json:"primary_channels"`
}

func TestPrimaryChannels(t *testing.T) {
	t.Parallel()

	// arrange
	os.Setenv("ENV", "staging")
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
			url := fmt.Sprintf(baseUrl+primaryChannelEndpoint, emIngress.Protocol, emIngress.Host, emIngress.APIServerPath, ch.Name)
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

				defer resp.Body.Close()

				data, err := json.Marshal(respValues)
				require.NoError(t, err)

				var channels PrimaryChannels
				err = json.Unmarshal(data, &channels)
				require.NoError(t, err)
				require.NotNil(t, channels)

				for _, v := range channels.PrimaryChannels {
					// arrange
					counterPartyURL := fmt.Sprintf(baseUrl+channelCounterparty, emIngress.Protocol, emIngress.Host, emIngress.APIServerPath, ch.Name, v.Counterparty)
					// act
					resp, err := client.Get(counterPartyURL)
					require.NoError(t, err)

					defer resp.Body.Close()

					var respValues map[string]interface{}
					utils.RespBodyToMap(resp.Body, &respValues, t)

					// expect a non empty data
					require.NotNil(t, respValues)

				}
			}
		})
	}
}
