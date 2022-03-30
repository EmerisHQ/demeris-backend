package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/allinbits/demeris-backend-models/cns"
	utils "github.com/allinbits/demeris-backend/test_utils"
)

const (
	primaryChannelsEndpoint = "chain/%s/primary_channels"
	primaryChannelskey      = "primary_channels"
)

func (suite *testCtx) TestPrimaryChannels() {
	for _, ch := range suite.Chains {
		suite.T().Run(ch.Name, func(t *testing.T) {
			// arrange
			url := suite.Client.BuildUrl(primaryChannelsEndpoint, ch.Name)
			// act
			resp, err := suite.Client.Get(url)
			suite.Require().NoError(err)

			defer resp.Body.Close()

			// assert
			if !ch.Enabled {
				suite.Require().Equal(http.StatusBadRequest, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))
			} else {
				suite.Require().Equal(http.StatusOK, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))

				var respValues map[string]interface{}
				utils.RespBodyToMap(resp.Body, &respValues, t)

				data, err := json.Marshal(respValues[primaryChannelskey])
				suite.Require().NoError(err)

				var channels []cns.DbStringMap
				err = json.Unmarshal(data, &channels)
				suite.Require().NoError(err)

				formattedChannels := make(map[string]string, len(channels))
				for _, channel := range channels {
					formattedChannels[channel["counterparty"]] = channel["channel_name"]
				}

				var payload map[string]interface{}
				err = json.Unmarshal(ch.Payload, &payload)
				suite.Require().NoError(err)

				data, err = json.Marshal(payload["primary_channel"])
				suite.Require().NoError(err)

				var expectedChannels map[string]string
				err = json.Unmarshal(data, &expectedChannels)
				suite.Require().NoError(err)

				suite.Require().Equal(expectedChannels, formattedChannels)
			}
		})
	}
}
