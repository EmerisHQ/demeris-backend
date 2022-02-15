package tests

import (
	"encoding/json"
	"fmt"
	"net/http"

	chainClient "github.com/allinbits/demeris-backend/chain_client"
	"github.com/allinbits/demeris-backend/models"
	utils "github.com/allinbits/demeris-backend/test_utils"
)

const (
	ChainNumbersEndpoint = "/chain/%s/numbers/%v"
)

func (suite *testCtx) TestGetChainNumbers() {
	suite.T().Parallel()

	for _, ch := range suite.clientChains {
		suite.Run(ch.Name, func() {
			var cc chainClient.Client
			err := json.Unmarshal(ch.Payload, &cc)
			suite.Require().NoError(err)
			cli := chainClient.GetClient(suite.T(), suite.Env, ch.Name, cc)
			suite.Require().NotNil(cli)

			hexAddress, err := cli.GetHexAddress(cc.Key)
			suite.Require().NoError(err)
			url := fmt.Sprintf(baseUrl+ChainNumbersEndpoint, suite.EmIngress.Protocol, suite.EmIngress.Host, suite.EmIngress.APIServerPath, ch.Name, hexAddress)
			// act
			resp, err := suite.Client.Get(url)
			suite.Require().NoError(err)

			if !cli.Enabled {
				suite.Require().Equal(http.StatusBadRequest, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))
				err = resp.Body.Close()
				suite.Require().NoError(err)
			} else {
				suite.Require().Equal(http.StatusOK, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))

				var respValues map[string]interface{}
				utils.RespBodyToMap(resp.Body, &respValues, suite.T())

				err = resp.Body.Close()
				suite.Require().NoError(err)
				suite.Require().NotNil(respValues)

				data, _ := json.Marshal(respValues["numbers"])
				suite.Require().NotNil(data)

				var row models.AuthRow
				err = json.Unmarshal(data, &row)
				suite.Require().NoError(err)
				suite.Require().NotNil(row)

				account, err := cli.AccountGet(cc.Key)
				suite.Require().NoError(err)

				suite.Require().Equal(ch.Name, row.ChainName)
				suite.Require().Equal(account.Address, row.Address)

				suite.Require().NotZero(row.AccountNumber)
				suite.Require().NotZero(row.SequenceNumber)
			}
		})
	}
}
