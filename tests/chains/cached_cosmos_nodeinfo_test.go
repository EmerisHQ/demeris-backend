package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	chainClient "github.com/allinbits/demeris-backend/chain_client"
	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	tmjson "github.com/tendermint/tendermint/libs/json"
)

const (
	cachedCosmosNodeEndpoint = "cached/cosmos/node_info"
)

func (suite *testCtx) TestCachedCosmosNodeinfo() {
	for _, ch := range suite.clientChains {
		suite.Run(ch.Name, func() {
			if ch.Name == "cosmos-hub" {
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
				var nodeInfo tmservice.GetNodeInfoResponse
				suite.Require().NoError(tmjson.Unmarshal(data, &nodeInfo))

				//get cosmos nodeinfo
				nodeInfoQuery := tmservice.NewServiceClient(cli.GetContext())
				nodeInfoRes, err := nodeInfoQuery.GetNodeInfo(context.Background(), &tmservice.GetNodeInfoRequest{})
				suite.Require().NoError(err)

				n1, err := tmjson.Marshal(nodeInfo)
				suite.Require().NoError(err)

				n2, err := tmjson.Marshal(nodeInfoRes)
				suite.Require().NoError(err)

				// match result
				suite.Require().Equal(string(n1), string(n2))
			}
		})
	}
}
