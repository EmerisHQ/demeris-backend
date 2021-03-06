package tests

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/emerishq/demeris-backend-models/cns"
	utils "github.com/emerishq/demeris-backend/test_utils"
)

const (
	primaryChannelsEndpoint = "chain/%s/primary_channels"
	primaryChannelskey      = "primary_channels"
)

func (suite *testCtx) TestPrimaryChannels() {
	for _, ch := range suite.Chains {
		suite.Run(ch.ChainName, func() {
			// arrange
			url := suite.Client.BuildUrl(primaryChannelsEndpoint, ch.ChainName)
			// act
			resp, err := suite.Client.Get(url)
			suite.Require().NoError(err)

			defer resp.Body.Close()

			// assert
			if !ch.Enabled {
				suite.Require().Equal(http.StatusBadRequest, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.ChainName, resp.StatusCode))
			} else {
				suite.Require().Equal(http.StatusOK, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.ChainName, resp.StatusCode))

				var respValues map[string]interface{}
				utils.RespBodyToMap(resp.Body, &respValues, suite.T())

				data, err := json.Marshal(respValues[primaryChannelskey])
				suite.Require().NoError(err)

				var channels []cns.DbStringMap
				err = json.Unmarshal(data, &channels)
				suite.Require().NoError(err)

				formattedChannels := make(cns.DbStringMap, len(channels))
				for _, channel := range channels {
					formattedChannels[channel["counterparty"]] = channel["channel_name"]
				}

				suite.Require().Equal(ch.PrimaryChannel, formattedChannels)
			}
		})
	}
}
