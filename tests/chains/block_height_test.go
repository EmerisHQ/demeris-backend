package tests

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	blockModels "github.com/allinbits/demeris-api-server/api/block"
	chainclient "github.com/allinbits/demeris-backend/chainclient"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
)

const (
	blockHeightEndpoint = "block_results?height=%d"
)

func (suite *testCtx) TestBlockHeight() {
	for _, ch := range suite.clientChains {
		if ch.ChainName != "cosmos-hub" {
			continue
		}
		suite.Run(ch.ChainName, func() {
			cli, _ := chainclient.GetClient(suite.Env, ch.ChainName, ch, suite.T().TempDir())
			suite.Require().NotNil(cli)
			if !cli.Enabled {
				return
			}

			//get block results from the cosmos node
			cRes, err := suite.Client.Get(ch.RPC + "/block_results")
			suite.Require().NoError(err)
			data, err := ioutil.ReadAll(cRes.Body)
			suite.Require().NoError(err)

			err = cRes.Body.Close()
			suite.Require().NoError(err)

			var cosmosBlock coretypes.ResultBlock
			err = json.Unmarshal(data, &cosmosBlock)
			suite.Require().NoError(err)
			suite.Require().NotNil(cosmosBlock.Block)

			//get block results from the env
			encodedStr := suite.Client.BuildUrl(blockHeightEndpoint, cosmosBlock.Block.Height)
			actualUrl, err := url.PathUnescape(encodedStr)
			suite.Require().NoError(err)

			// act
			resp, err := suite.Client.Get(actualUrl)
			suite.Require().NoError(err)

			suite.Require().Equal(http.StatusOK, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.ChainName, resp.StatusCode))

			blockData, err := ioutil.ReadAll(resp.Body)
			suite.Require().NoError(err)

			err = resp.Body.Close()
			suite.Require().NoError(err)

			var block blockModels.BlockHeightResp
			err = json.Unmarshal(blockData, block)
			suite.Require().NoError(err)

			suite.Require().Equal(cosmosBlock, block.Result)
		})
	}
}
