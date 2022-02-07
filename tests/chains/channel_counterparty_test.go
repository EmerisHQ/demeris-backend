package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	utils "github.com/allinbits/demeris-backend/test_utils"
)

const (
	primaryChannelkey   = "primary_channel"
	channelCounterparty = "/chain/%s/primary_channel/%s"
)

func (suite *testCtx) TestPrimaryChannelCounterparty() {
	suite.T().Parallel()

	for _, ch := range suite.chains {
		suite.T().Run(ch.Name, func(t *testing.T) {
			if !ch.Enabled {
				// checking /chain/XXX/primary_channel/ZZZ returns 400 if chain disabled
				for _, otherChains := range suite.chains {
					if otherChains.Name != ch.Name {
						// arrange
						counterPartyURL := fmt.Sprintf(baseUrl+channelCounterparty, suite.emIngress.Protocol, suite.emIngress.Host, suite.emIngress.APIServerPath, ch.Name, otherChains.Name)
						// act
						resp, err := suite.client.Get(counterPartyURL)
						suite.NoError(err)

						suite.Equal(http.StatusBadRequest, resp.StatusCode, fmt.Sprintf("Chain %s Channel %s HTTP code %d", ch.Name, otherChains.Name, resp.StatusCode))
					}
				}
			} else {
				var payload map[string]interface{}
				err := json.Unmarshal(ch.Payload, &payload)
				suite.NoError(err)

				data, err := json.Marshal(payload[primaryChannelkey])
				suite.NoError(err)

				var expectedChannels map[string]string
				err = json.Unmarshal(data, &expectedChannels)
				suite.NoError(err)

				// test for existing channels
				for counterParty, channel_name := range expectedChannels {
					// arrange
					counterPartyURL := fmt.Sprintf(baseUrl+channelCounterparty, suite.emIngress.Protocol, suite.emIngress.Host, suite.emIngress.APIServerPath, ch.Name, counterParty)
					// act
					resp, err := suite.client.Get(counterPartyURL)
					suite.NoError(err)

					suite.Equal(http.StatusOK, resp.StatusCode, fmt.Sprintf("Chain %s Channel %s HTTP code %d", ch.Name, counterParty, resp.StatusCode))

					defer resp.Body.Close()

					var respValues map[string]interface{}
					utils.RespBodyToMap(resp.Body, &respValues, t)

					expectedChannelsFormatted := map[string]interface{}{
						"counterparty": counterParty,
						"channel_name": channel_name,
					}
					suite.Equal(expectedChannelsFormatted, respValues[primaryChannelkey])
				}

				// test for non-existing channels
				for _, otherChains := range suite.chains {
					if _, ok := expectedChannels[otherChains.Name]; !ok && otherChains.Name != ch.Name {
						// arrange
						counterPartyURL := fmt.Sprintf(baseUrl+channelCounterparty, suite.emIngress.Protocol, suite.emIngress.Host, suite.emIngress.APIServerPath, ch.Name, otherChains.Name)
						// act
						resp, err := suite.client.Get(counterPartyURL)
						suite.NoError(err)

						suite.Equal(http.StatusBadRequest, resp.StatusCode, fmt.Sprintf("Chain %s Channel %s HTTP code %d", ch.Name, otherChains.Name, resp.StatusCode))
					}
				}
			}
		})
	}
}
