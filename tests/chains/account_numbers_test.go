package tests

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	models "github.com/allinbits/demeris-backend-models/api"
	chainClient "github.com/allinbits/demeris-backend/chain_client"
)

const (
	accountNumbersEndpoint = "account/%v/numbers"
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

			url := suite.Client.BuildUrl(accountNumbersEndpoint, hex.EncodeToString(hexAddress))
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

			data, err := ioutil.ReadAll(resp.Body)
			suite.Require().NoError(err)

			err = resp.Body.Close()
			suite.Require().NoError(err)

			var numbers models.AccountNumbersResponse
			suite.Require().NoError(json.Unmarshal(data, &numbers))
			suite.Require().NotEmpty(numbers.Numbers)

		})
	}
}
