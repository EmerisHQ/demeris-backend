package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	utils "github.com/allinbits/demeris-backend/test_utils"
)

const (
	primaryChannelkey   = "primary_channel"
	channelCounterparty = "/chain/%s/primary_channel/%s"
)

func TestPrimaryChannelCounterparty(t *testing.T) {
	t.Parallel()

	for _, ch := range testCtx.chains {
		t.Run(ch.Name, func(t *testing.T) {
			if !ch.Enabled {
				// checking /chain/XXX/primary_channel/ZZZ returns 400 if chain disabled
				for _, otherChains := range testCtx.chains {
					if otherChains.Name != ch.Name {
						// arrange
						counterPartyURL := fmt.Sprintf(baseUrl+channelCounterparty, testCtx.emIngress.Protocol, testCtx.emIngress.Host, testCtx.emIngress.APIServerPath, ch.Name, otherChains.Name)
						// act
						resp, err := testCtx.client.Get(counterPartyURL)
						require.NoError(t, err)

						require.Equal(t, http.StatusBadRequest, resp.StatusCode, fmt.Sprintf("Chain %s Channel %s HTTP code %d", ch.Name, otherChains.Name, resp.StatusCode))
					}
				}
			} else {
				var payload map[string]interface{}
				err := json.Unmarshal(ch.Payload, &payload)
				require.NoError(t, err)

				data, err := json.Marshal(payload[primaryChannelkey])
				require.NoError(t, err)

				var expectedChannels map[string]string
				err = json.Unmarshal(data, &expectedChannels)
				require.NoError(t, err)

				// test for existing channels
				for counterParty, channel_name := range expectedChannels {
					// arrange
					counterPartyURL := fmt.Sprintf(baseUrl+channelCounterparty, testCtx.emIngress.Protocol, testCtx.emIngress.Host, testCtx.emIngress.APIServerPath, ch.Name, counterParty)
					// act
					resp, err := testCtx.client.Get(counterPartyURL)
					require.NoError(t, err)

					require.Equal(t, http.StatusOK, resp.StatusCode, fmt.Sprintf("Chain %s Channel %s HTTP code %d", ch.Name, counterParty, resp.StatusCode))

					defer resp.Body.Close()

					var respValues map[string]interface{}
					utils.RespBodyToMap(resp.Body, &respValues, t)

					expectedChannelsFormatted := map[string]interface{}{
						"counterparty": counterParty,
						"channel_name": channel_name,
					}
					require.Equal(t, expectedChannelsFormatted, respValues[primaryChannelkey])
				}

				// test for non-existing channels
				for _, otherChains := range testCtx.chains {
					if _, ok := expectedChannels[otherChains.Name]; !ok && otherChains.Name != ch.Name {
						// arrange
						counterPartyURL := fmt.Sprintf(baseUrl+channelCounterparty, testCtx.emIngress.Protocol, testCtx.emIngress.Host, testCtx.emIngress.APIServerPath, ch.Name, otherChains.Name)
						// act
						resp, err := testCtx.client.Get(counterPartyURL)
						require.NoError(t, err)

						require.Equal(t, http.StatusBadRequest, resp.StatusCode, fmt.Sprintf("Chain %s Channel %s HTTP code %d", ch.Name, otherChains.Name, resp.StatusCode))
					}
				}
			}
		})
	}
}
