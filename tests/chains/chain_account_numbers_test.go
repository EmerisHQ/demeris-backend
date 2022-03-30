package tests

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	chainModels "github.com/allinbits/demeris-api-server/api/chains"
	chainClient "github.com/allinbits/demeris-backend/chain_client"
)

const (
	chainNumbersEndpoint = "chain/%s/numbers/%v"
)

func (suite *testCtx) TestGetChainNumbers() {
	for _, ch := range suite.clientChains {
		suite.Run(ch.Name, func() {
			var cc chainClient.ChainClient
			err := json.Unmarshal(ch.Payload, &cc)
			suite.Require().NoError(err)
			cli, err := chainClient.GetClient(suite.Env, ch.Name, cc, suite.T().TempDir())
			suite.Require().NoError(err)
			suite.Require().NotNil(cli)

			accAddr, err := cli.GetAccAddress(cc.Key)
			suite.Require().NoError(err)

			url := suite.Client.BuildUrl(chainNumbersEndpoint, ch.Name, hex.EncodeToString(accAddr))
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
			suite.Require().NotNil(data)

			err = resp.Body.Close()
			suite.Require().NoError(err)

			var row chainModels.NumbersResponse
			err = json.Unmarshal(data, &row)
			suite.Require().NoError(err)
			suite.Require().NotNil(row)

			account, err := cli.AccountGet(cc.Key)
			suite.Require().NoError(err)

			suite.Require().Equal(ch.Name, row.Numbers.ChainName)
			suite.Require().Equal(account.Address, row.Numbers.Address)
		})
	}
}
