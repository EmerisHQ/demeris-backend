package tests

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/allinbits/demeris-api-server/api/block"
	chainClient "github.com/allinbits/demeris-backend/chain_client"
)

const (
	blockHeightEndpoint = "block_results?height=%d"
)

func (suite *testCtx) TestBlockHeight() {
	for _, ch := range suite.clientChains {
		if ch.Name == "cosmos-hub" {
			suite.Run(ch.Name, func() {
				var cc chainClient.Client
				err := json.Unmarshal(ch.Payload, &cc)
				suite.Require().NoError(err)
				cli := chainClient.GetClient(suite.T(), suite.Env, ch.Name, cc)
				suite.Require().NotNil(cli)
				if !cli.Enabled {
					return
				}

				//get block results from the cosmos node
				cRes, err := suite.Client.Get(cc.RPC + "/block_results")
				suite.Require().NoError(err)
				data, err := ioutil.ReadAll(cRes.Body)
				suite.Require().NoError(err)

				err = cRes.Body.Close()
				suite.Require().NoError(err)

				var cosmosBlock block.BlockHeightResp
				err = json.Unmarshal(data, &cosmosBlock)
				suite.Require().NoError(err)

				//get block results from the env
				encodedStr := suite.Client.BuildUrl(blockHeightEndpoint, cosmosBlock.Result.Block.Height)
				actualUrl, err := url.PathUnescape(encodedStr)
				suite.Require().NoError(err)

				// act
				resp, err := suite.Client.Get(actualUrl)
				suite.Require().NoError(err)

				suite.Require().Equal(http.StatusOK, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))

				blockData, err := ioutil.ReadAll(resp.Body)
				suite.Require().NoError(err)

				err = resp.Body.Close()
				suite.Require().NoError(err)

				var block interface{}
				err = json.Unmarshal(blockData, block)
				suite.Require().NoError(err)

				suite.Require().Equal(cosmosBlock, block)
			})
		}
	}
}
