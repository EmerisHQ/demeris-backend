package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	chainClient "github.com/allinbits/demeris-backend/chain_client"
	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
)

const (
	blockHeightEndpoint = "block_results?height=%v"
)

func (suite *testCtx) TestBlockHeight() {
	for _, ch := range suite.clientChains {
		suite.Run(ch.Name, func() {
			if ch.Name == "cosmos-hub" {
				var cc chainClient.Client
				err := json.Unmarshal(ch.Payload, &cc)
				suite.Require().NoError(err)
				cli := chainClient.GetClient(suite.T(), suite.Env, ch.Name, cc)
				suite.Require().NotNil(cli)
				if !cli.Enabled {
					return
				}

				//Get latestblock suing client from tmservice
				latestBlockQuery := tmservice.NewServiceClient(cli.GetContext())
				latestBlockRes, err := latestBlockQuery.GetLatestBlock(context.Background(), &tmservice.GetLatestBlockRequest{})
				suite.Require().NoError(err)

				fmt.Println("block height......", latestBlockRes.Block.Header.Height)

				//get block results from the endpoint
				url := suite.Client.BuildUrl(blockHeightEndpoint, latestBlockRes.Block.Header.Height)
				// act
				resp, err := suite.Client.Get(url)
				suite.Require().NoError(err)

				suite.Require().Equal(http.StatusOK, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))

				data, err := ioutil.ReadAll(resp.Body)
				suite.Require().NoError(err)

				err = resp.Body.Close()
				suite.Require().NoError(err)

				var block interface{}
				err = json.Unmarshal(data, block)
				suite.Require().NoError(err)

			}
		})
	}
}
