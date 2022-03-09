package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	chainClient "github.com/allinbits/demeris-backend/chain_client"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

const (
	stakingParamsEndpoint = "chain/%v/staking/params"
)

func (suite *testCtx) TestStakingParams() {
	for _, ch := range suite.clientChains {
		suite.Run(ch.Name, func() {
			var cc chainClient.Client
			err := json.Unmarshal(ch.Payload, &cc)
			suite.Require().NoError(err)

			cli := chainClient.GetClient(suite.T(), suite.Env, ch.Name, cc)

			// arrange
			url := suite.Client.BuildUrl(stakingParamsEndpoint, ch.Name)
			// act
			resp, err := suite.Client.Get(url)
			suite.Require().NoError(err)
			// assert
			if !cli.Enabled {
				suite.Require().Equal(http.StatusBadRequest, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))
			} else {
				suite.Require().Equal(http.StatusOK, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))

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
