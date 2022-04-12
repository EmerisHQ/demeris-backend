package tests

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	chainclient "github.com/allinbits/demeris-backend/chain_client"
	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	tmjson "github.com/tendermint/tendermint/libs/json"
)

const (
	cachedCosmosNodeEndpoint = "cached/cosmos/node_info"
)

func (suite *testCtx) TestCachedCosmosNodeinfo() {
	for _, ch := range suite.clientChains {
		if ch.ChainName != "cosmos-hub" {
			continue
		}
		suite.Run(ch.ChainName, func() {
			cli, err := chainclient.GetClient(suite.Env, ch.ChainName, ch, suite.T().TempDir())
			suite.Require().NoError(err)
			suite.Require().NotNil(cli)

			url := suite.Client.BuildUrl(cachedCosmosNodeEndpoint)
			// act
			resp, err := suite.Client.Get(url)
			suite.Require().NoError(err)

			suite.Require().Equal(http.StatusOK, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.ChainName, resp.StatusCode))

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

			// match result
			suite.Require().Equal(nodeInfo.DefaultNodeInfo.Network, nodeInfoRes.DefaultNodeInfo.Network)
		})
	}
}
