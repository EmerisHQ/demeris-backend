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

func TestPrimaryChannelCounterparty(t *testing.T) {
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
			if !ch.Enabled {
				// checking /chain/XXX/primary_channel/ZZZ returns 400 if chain disabled
				for _, otherChains := range chains {
					if otherChains.Name != ch.Name {
						// arrange
						counterPartyURL := fmt.Sprintf(baseUrl+channelCounterparty, emIngress.Protocol, emIngress.Host, emIngress.APIServerPath, ch.Name, otherChains.Name)
						// act
						resp, err := client.Get(counterPartyURL)
						require.NoError(t, err)

						require.Equal(t, http.StatusBadRequest, resp.StatusCode, fmt.Sprintf("Chain %s Channel %s HTTP code %d", ch.Name, otherChains.Name, resp.StatusCode))
					}
				}
			} else {
				var payload map[string]interface{}
				err := json.Unmarshal(ch.Payload, &payload)
				require.NoError(t, err)

				data, err := json.Marshal(payload["primary_channel"])
				require.NoError(t, err)

				var expectedChannels map[string]string
				err = json.Unmarshal(data, &expectedChannels)
				require.NoError(t, err)

				// test for existing channels
				for counterParty, channel_name := range expectedChannels {
					// arrange
					counterPartyURL := fmt.Sprintf(baseUrl+channelCounterparty, emIngress.Protocol, emIngress.Host, emIngress.APIServerPath, ch.Name, counterParty)
					// act
					resp, err := client.Get(counterPartyURL)
					require.NoError(t, err)

					require.Equal(t, http.StatusOK, resp.StatusCode, fmt.Sprintf("Chain %s Channel %s HTTP code %d", ch.Name, counterParty, resp.StatusCode))

					defer resp.Body.Close()

					var respValues map[string]interface{}
					utils.RespBodyToMap(resp.Body, &respValues, t)

					expectedChannelsFormatted := map[string]interface{}{
						"counterparty": counterParty,
						"channel_name": channel_name,
					}
					require.Equal(t, expectedChannelsFormatted, respValues["primary_channel"])
				}

				// test for non-existing channels
				for _, otherChains := range chains {
					if _, ok := expectedChannels[otherChains.Name]; !ok && otherChains.Name != ch.Name {
						// arrange
						counterPartyURL := fmt.Sprintf(baseUrl+channelCounterparty, emIngress.Protocol, emIngress.Host, emIngress.APIServerPath, ch.Name, otherChains.Name)
						// act
						resp, err := client.Get(counterPartyURL)
						require.NoError(t, err)

						require.Equal(t, http.StatusBadRequest, resp.StatusCode, fmt.Sprintf("Chain %s Channel %s HTTP code %d", ch.Name, otherChains.Name, resp.StatusCode))
					}
				}
			}
		})
	}
}
