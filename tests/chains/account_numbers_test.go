package tests

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	apiServer "github.com/allinbits/demeris-api-server/api/chains"
	chainClient "github.com/allinbits/demeris-backend/chain_client"
	utils "github.com/allinbits/demeris-backend/test_utils"
)

const (
	AccountNumbersEndpoint = "account/%v/numbers"
)

func (suite *testCtx) TestGetAccountNumbers() {
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

			urlPattern := strings.Join([]string{baseUrl, AccountNumbersEndpoint}, "")
			url := fmt.Sprintf(urlPattern, suite.EmIngress.Protocol, suite.EmIngress.Host, suite.EmIngress.APIServerPath, hex.EncodeToString(hexAddress))
			// act
			resp, err := suite.Client.Get(url)
			suite.Require().NoError(err)

			if !cli.Enabled {
				suite.Require().Equal(http.StatusBadRequest, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))
				err = resp.Body.Close()
				suite.Require().NoError(err)

				return
			}
			suite.Require().Equal(http.StatusOK, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.Name, resp.StatusCode))

			var respValues map[string]interface{}
			utils.RespBodyToMap(resp.Body, &respValues, suite.T())

			err = resp.Body.Close()
			suite.Require().NoError(err)
			suite.Require().NotNil(respValues)

			data, err := json.Marshal(respValues["numbers"])
			suite.Require().NoError(err)
			suite.Require().NotNil(data)

			var row []apiServer.NumbersResponse
			err = json.Unmarshal(data, &row)
			suite.Require().NoError(err)
			suite.Require().NotNil(row)

		})
	}
}
