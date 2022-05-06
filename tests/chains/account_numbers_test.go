package tests

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	models "github.com/emerishq/demeris-api-server/api/account"
	chainclient "github.com/emerishq/demeris-backend/chainclient"
)

const (
	accountNumbersEndpoint = "account/%v/numbers"
)

func (suite *testCtx) TestGetAccountNumbers() {
	// emeris/api-server-6bf5b6655c-fkzkt[api-server]: 2022-05-03T10:07:13.056Z        ERROR   router/router.go:143    numbers: cannot query nodes auth for addresses: cannot query chains, unable to get account numbers, cannot query account numbers, rpc error: code = Unavailable desc = connection error: desc = "transport: Error while dialing dial tcp 10.16.190.92:9090: connect: connection refused" {"int_correlation_id": "7aaf4257-6c80-4b82-b96a-75c3d8c44cbf", "address": "3b2db11d20750d2f67ad818e9b2055614682664d", "address": "3b2db11d20750d2f67ad818e9b2055614682664d", "error": "numbers: cannot query nodes auth for addresses: cannot query chains, unable to get account numbers, cannot query account numbers, rpc error: code = Unavailable desc = connection error: desc = \"transport: Error while dialing dial tcp 10.16.190.92:9090: connect: connection refused\""}
	// suite.T().Skip("skip: api-server errors, full error in the comment of this test")

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
