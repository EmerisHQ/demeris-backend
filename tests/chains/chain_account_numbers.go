package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	chainClient "github.com/allinbits/demeris-backend/chain_client"
	"github.com/allinbits/demeris-backend/models"
	utils "github.com/allinbits/demeris-backend/test_utils"
	"github.com/stretchr/testify/require"
)

const (
	ChainNumbersEndpoint = "/chain/%s/numbers/%v"
)

func (suite *testCtx) TestGetChainNumbers(t *testing.T) {
	suite.T().Parallel()

	for _, ch := range suite.clientChains {
		suite.Run(ch.Name, func() {
			var cc chainClient.Client
			err := json.Unmarshal(ch.Payload, &cc)
			suite.Require().NoError(err)
			cli := chainClient.GetClient(suite.T(), suite.env, ch.Name, cc)

			hexAddress, err := cc.GetHexAddress(ch.Name)
			require.NoError(t, err)
			url := fmt.Sprintf(baseUrl+ChainNumbersEndpoint, suite.emIngress.Protocol, suite.emIngress.Host, suite.emIngress.APIServerPath, ch.Name, hexAddress)
			// act
			resp, err := suite.client.Get(url)
			suite.Require().NoError(err)

			if !cli.Enabled {
				suite.Require().Equal(http.StatusBadRequest, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))
				err = resp.Body.Close()
				suite.Require().NoError(err)
			} else {
				suite.Require().Equal(http.StatusBadRequest, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))

				var respValues map[string]interface{}
				utils.RespBodyToMap(resp.Body, &respValues, t)

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

				require.Equal(t, cli.GetContext().ChainID, row.ChainName)
				require.Equal(t, account.Address, row.Address)

				require.NotZero(t, row.AccountNumber)
				require.NotZero(t, row.SequenceNumber)
			}
		})
	}
}
