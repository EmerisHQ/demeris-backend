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
	suite.T().Parallel()

	for _, ch := range suite.Chains {
		suite.T().Run(ch.Name, func(t *testing.T) {
			// arrange
			url := fmt.Sprintf(baseUrl+primaryChannelsEndpoint, suite.EmIngress.Protocol, suite.EmIngress.Host, suite.EmIngress.APIServerPath, ch.Name)
			// act
			resp, err := suite.Client.Get(url)
			suite.NoError(err)

			defer resp.Body.Close()

			// assert
			if !ch.Enabled {
				suite.Equal(http.StatusBadRequest, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))
			} else {
				suite.Equal(http.StatusOK, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))

				var respValues map[string]interface{}
				utils.RespBodyToMap(resp.Body, &respValues, t)

				data, err := json.Marshal(respValues[primaryChannelskey])
				suite.NoError(err)

				var channels []cns.DbStringMap
				err = json.Unmarshal(data, &channels)
				suite.NoError(err)

				formattedChannels := make(map[string]string, len(channels))
				for _, channel := range channels {
					formattedChannels[channel["counterparty"]] = channel["channel_name"]
				}

				var payload map[string]interface{}
				err = json.Unmarshal(ch.Payload, &payload)
				suite.NoError(err)

				data, err = json.Marshal(payload["primary_channel"])
				suite.NoError(err)

				var expectedChannels map[string]string
				err = json.Unmarshal(data, &expectedChannels)
				suite.NoError(err)

				suite.Equal(expectedChannels, formattedChannels)
			}
		})
	}
}
