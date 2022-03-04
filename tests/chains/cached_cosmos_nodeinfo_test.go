package tests

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/allinbits/demeris-backend-models/cns"
	chainClient "github.com/allinbits/demeris-backend/chain_client"
)

const (
	cachedCosmosNodeEndpoint = "/cached/cosmos/node_info"
)

func (suite *testCtx) TestCachedCosmosNodeinfo() {
	for _, ch := range suite.clientChains {
		suite.Run(ch.Name, func() {
			var cc chainClient.Client
			err := json.Unmarshal(ch.Payload, &cc)
			suite.Require().NoError(err)
			cli := chainClient.GetClient(suite.T(), suite.Env, ch.Name, cc)
			suite.Require().NotNil(cli)

			url := suite.Client.BuildUrl(cachedCosmosNodeEndpoint)
			// act
			resp, err := suite.Client.Get(url)
			suite.Require().NoError(err)

			suite.Require().Equal(http.StatusOK, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))

			data, err := ioutil.ReadAll(resp.Body)
			suite.Require().NoError(err)

			err = resp.Body.Close()
			suite.Require().NoError(err)

			if !cli.Enabled {
				return
			}

			var nodeInfo cns.NodeInfo
			suite.Require().NoError(json.Unmarshal(data, &nodeInfo))

			if ch.Name == "cosmos-hub" {
				var info cns.NodeInfo
				err = json.Unmarshal(ch.Payload, &info)
				suite.Require().NoError(err)

				suite.Require().Equal(nodeInfo, info)
			}
		})
	}
}
