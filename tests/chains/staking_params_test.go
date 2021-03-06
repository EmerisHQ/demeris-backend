package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	chainclient "github.com/emerishq/demeris-backend/chainclient"
)

const (
	stakingParamsEndpoint = "chain/%v/staking/params"
)

func (suite *testCtx) TestStakingParams() {
	for _, ch := range suite.clientChains {
		suite.Run(ch.ChainName, func() {
			cli, err := chainclient.GetClient(suite.Env, ch.ChainName, ch, suite.T().TempDir())
			suite.Require().NoError(err)

			// arrange
			url := suite.Client.BuildUrl(stakingParamsEndpoint, ch.ChainName)
			// act
			resp, err := suite.Client.Get(url)
			suite.Require().NoError(err)
			// assert
			if !cli.Enabled {
				suite.Require().Equal(http.StatusBadRequest, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.ChainName, resp.StatusCode))
			} else {
				suite.Require().Equal(http.StatusOK, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.ChainName, resp.StatusCode))

				data, err := ioutil.ReadAll(resp.Body)
				suite.Require().NoError(err)

				var params stakingtypes.QueryParamsResponse
				suite.Require().NoError(json.Unmarshal(data, &params))
				suite.Require().NotEmpty(params.Params)

				nodeParams, err := stakingtypes.NewQueryClient(cli.GetContext()).Params(context.Background(), &stakingtypes.QueryParamsRequest{})
				suite.Require().NoError(err)
				suite.Require().NotEmpty(nodeParams)

				suite.Require().Equal(params.Params, nodeParams.Params)
			}

			err = resp.Body.Close()
			suite.Require().NoError(err)
		})
	}
}
