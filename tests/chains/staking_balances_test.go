package tests

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	api "github.com/allinbits/demeris-api-server/api/account"
	chainClient "github.com/allinbits/demeris-backend/chain_client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

const (
	stakingBalanceEndpoint = "account/%s/stakingbalances"
)

func (suite *testCtx) TestStakingBalance() {
	for _, ch := range suite.clientChains {
		suite.Run(ch.ChainName, func() {
			cli, err := chainClient.GetClient(suite.Env, ch.ChainName, ch, suite.T().TempDir())
			suite.Require().NoError(err)

			if !cli.Enabled {
				return
			}

			address, err := cli.GetAccAddress(ch.Key)
			suite.Require().NoError(err)

			expectedDelegations, err := cli.GetDelegations(address.String())
			suite.Require().NoError(err)

			// assert
			if expectedDelegations == nil || len(expectedDelegations) == 0 {
				validators, err := cli.GetUnbondedValidators()
				suite.Require().NoError(err)
				suite.Require().NotEmpty(validators, "no validators found to do stake tx")

				// check balance
				balance, err := cli.GetAccountBalances(address.String(), cli.Denom)
				suite.Require().NoError(err)
				suite.Require().True(balance.Amount.GT(sdk.NewInt(100)), "not enough balance in given account to perform tx")

				// get validator address
				valAddr, err := sdk.ValAddressFromBech32(validators[0].OperatorAddress)
				suite.Require().NoError(err)

				// perform unbonding transaction
				msg := stakingtypes.NewMsgDelegate(address, valAddr, sdk.NewCoin(cli.Denom, sdk.NewInt(100)))
				_, err = cli.Broadcast(ch.Key, cli.GetContext(), msg)
				suite.Require().NoError(err)

				// wait few sec to confirm
				time.Sleep(time.Second * 10)

				expectedDelegations, err = cli.GetDelegations(address.String())
				suite.Require().NoError(err)
			}

			// arrange
			url := suite.Client.BuildUrl(stakingBalanceEndpoint, hex.EncodeToString(address))
			// act
			resp, err := suite.Client.Get(url)
			suite.Require().NoError(err)
			suite.Require().Equal(http.StatusOK, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.ChainName, resp.StatusCode))

			data, err := ioutil.ReadAll(resp.Body)
			suite.Require().NoError(err)

			err = resp.Body.Close()
			suite.Require().NoError(err)

			var delegations api.StakingBalancesResponse
			suite.Require().NoError(json.Unmarshal(data, &delegations))
			suite.Require().NotEmpty(delegations.StakingBalances)

			// assert
			for _, delegation := range delegations.StakingBalances {
				// check for same chain name
				if delegation.ChainName == ch.ChainName {
					validatorFound := false
					for _, expected := range expectedDelegations {
						// get hex validator address
						valAddr, err := sdk.ValAddressFromBech32(expected.Delegation.ValidatorAddress)
						suite.Require().NoError(err)

						if delegation.ValidatorAddress == hex.EncodeToString(valAddr) {
							validatorFound = true
							delegationAmount, err := strconv.ParseFloat(delegation.Amount, 64)
							suite.Require().NoError(err)
							suite.Require().Equal(expected.Balance.Amount.Int64(), int64(delegationAmount))
						}
					}
					suite.Require().True(validatorFound)
				}
			}
		})
	}
}
