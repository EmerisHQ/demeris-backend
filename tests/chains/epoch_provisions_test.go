package tests

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
)

const (
	mintEpochProvisionsEndpoint = "chain/%s/mint/epoch_provisions"
)

func (suite *testCtx) TestEpochProvisions() {
	for _, ch := range suite.Chains {
		if ch.Name != "osmosis" {
			continue
		}
		suite.T().Run(ch.Name, func(t *testing.T) {

			// arrange
			url := suite.Client.BuildUrl(mintEpochProvisionsEndpoint, ch.Name)
			// act
			resp, err := suite.Client.Get(url)
			suite.Require().NoError(err)

			defer resp.Body.Close()

			// assert
			suite.Require().Equal(http.StatusOK, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))

			data, err := ioutil.ReadAll(resp.Body)
			suite.Require().NoError(err)

			var provisions json.RawMessage
			suite.Require().NoError(json.Unmarshal(data, &provisions))
			//expect a non empty data
			suite.Require().NotEmpty(provisions)
		})
	}
}
