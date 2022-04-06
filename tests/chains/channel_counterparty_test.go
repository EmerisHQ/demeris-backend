package tests

import (
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
	for _, ch := range suite.Chains {
		suite.T().Run(ch.ChainName, func(t *testing.T) {
			if !ch.Enabled {
				// checking /chain/XXX/primary_channel/ZZZ returns 400 if chain disabled
				for _, otherChains := range suite.Chains {
					if otherChains.ChainName != ch.ChainName {
						// arrange
						counterPartyURL := suite.Client.BuildUrl(channelCounterparty, ch.ChainName, otherChains.ChainName)
						// act
						resp, err := suite.Client.Get(counterPartyURL)
						suite.Require().NoError(err)

						suite.Require().Equal(http.StatusBadRequest, resp.StatusCode, fmt.Sprintf("Chain %s Channel %s HTTP code %d", ch.ChainName, otherChains.ChainName, resp.StatusCode))
					}
				}
			} else {
				expectedChannels := ch.PrimaryChannel

				// test for existing channels
				for counterParty, channel_name := range expectedChannels {
					// arrange
					counterPartyURL := suite.Client.BuildUrl(channelCounterparty, ch.ChainName, counterParty)
					// act
					resp, err := suite.Client.Get(counterPartyURL)
					suite.Require().NoError(err)

					suite.Require().Equal(http.StatusOK, resp.StatusCode, fmt.Sprintf("Chain %s Channel %s HTTP code %d", ch.ChainName, counterParty, resp.StatusCode))

					defer resp.Body.Close()

					var respValues map[string]interface{}
					utils.RespBodyToMap(resp.Body, &respValues, t)

					expectedChannelsFormatted := map[string]interface{}{
						"counterparty": counterParty,
						"channel_name": channel_name,
					}
					suite.Require().Equal(expectedChannelsFormatted, respValues[primaryChannelkey])
				}

				// test for non-existing channels
				for _, otherChains := range suite.Chains {
					if _, ok := expectedChannels[otherChains.ChainName]; !ok && otherChains.ChainName != ch.ChainName {
						// arrange
						counterPartyURL := suite.Client.BuildUrl(channelCounterparty, ch.ChainName, otherChains.ChainName)
						// act
						resp, err := suite.Client.Get(counterPartyURL)
						suite.Require().NoError(err)

						suite.Require().Equal(http.StatusBadRequest, resp.StatusCode, fmt.Sprintf("Chain %s Channel %s HTTP code %d", ch.ChainName, otherChains.ChainName, resp.StatusCode))
					}
				}
			}
		})
	}
}
