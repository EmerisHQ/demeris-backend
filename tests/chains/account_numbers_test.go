package tests

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	models "github.com/allinbits/demeris-api-server/api/account"
	chainclient "github.com/allinbits/demeris-backend/chainclient"
)

const (
	accountNumbersEndpoint = "account/%v/numbers"
)

func (suite *testCtx) TestGetAccountNumbers() {
	for _, ch := range suite.clientChains {
		suite.Run(ch.ChainName, func() {
			cli, err := chainclient.GetClient(suite.Env, ch.ChainName, ch, suite.T().TempDir())
			suite.Require().NoError(err)
			suite.Require().NotNil(cli)

			hexAddress, err := cli.GetAccAddress(ch.Key)
			suite.Require().NoError(err)

			url := suite.Client.BuildUrl(accountNumbersEndpoint, hex.EncodeToString(hexAddress))
			// act
			resp, err := suite.Client.Get(url)
			suite.Require().NoError(err)

			suite.Require().Equal(http.StatusOK, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.ChainName, resp.StatusCode))

			data, err := ioutil.ReadAll(resp.Body)
			suite.Require().NoError(err)

			err = resp.Body.Close()
			suite.Require().NoError(err)

			var numbers models.NumbersResponse
			suite.Require().NoError(json.Unmarshal(data, &numbers))

			if !cli.Enabled {
				suite.Require().Empty(numbers.Numbers)
				return
			}

			suite.Require().NotEmpty(numbers.Numbers)

			// get account information
			account, err := cli.AccountGet(ch.Key)
			suite.Require().NoError(err)

			// query account numbers from cli
			accNum, err := cli.GetAccountInfo(account.Address)
			suite.Require().NoError(err)

			// comapre account and sequence numbers
			for _, v := range numbers.Numbers {
				if v.ChainName == ch.ChainName {
					suite.Require().Equal(accNum.GetAccountNumber(), v.AccountNumber)
					suite.Require().Equal(accNum.GetSequence(), v.SequenceNumber)
					return
				}
			}
		})
	}
}
