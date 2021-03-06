package tests

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	api "github.com/emerishq/demeris-api-server/api/account"
	chainclient "github.com/emerishq/demeris-backend/chainclient"
)

const (
	unbondingDelegationsEndpoint = "account/%s/unbondingdelegations"
)

func (suite *testCtx) TestUnbondingDelegations() {
	suite.T().Skip("skip: we cannot unbond stake from validator itself, we need a different delegator")

	for _, ch := range suite.clientChains {
		suite.Run(ch.ChainName, func() {
			cli, err := chainclient.GetClient(suite.Env, ch.ChainName, ch, suite.T().TempDir())
			suite.Require().NoError(err)
			if !cli.Enabled {
				return
			}

			address, err := cli.GetAccAddress(ch.Key)
			suite.Require().NoError(err)

			expectedUndelegations, err := cli.GetUnbondingDelegations(address.String())
			suite.Require().NoError(err)

			// assert
			if expectedUndelegations == nil || len(expectedUndelegations) == 0 {
				validators, err := cli.GetBondedValidators()
				suite.Require().NoError(err)
				suite.Require().NotEmpty(validators, "no validators found to do unbond tx")

				// check balance
				balance, err := cli.GetAccountBalances(address.String(), cli.Denom)
				suite.Require().NoError(err)
				suite.Require().True(balance.Amount.GT(sdk.NewInt(100)), "not enough balance in given account to perform tx")

				// get validator address
				valAddr, err := sdk.ValAddressFromBech32(validators[0].OperatorAddress)
				suite.Require().NoError(err)

				// perform unbonding transaction
				msg := stakingtypes.NewMsgUndelegate(address, valAddr, sdk.NewCoin(cli.Denom, sdk.NewInt(100)))
				_, err = cli.Broadcast(ch.Key, address, cli.GetContext(), msg)
				suite.Require().NoError(err)

				// wait few sec to confirm
				time.Sleep(time.Second * 10)

				expectedUndelegations, err = cli.GetUnbondingDelegations(address.String())
				suite.Require().NoError(err)
			}

			// arrange
			url := suite.Client.BuildUrl(unbondingDelegationsEndpoint, hex.EncodeToString(address))
			// act
			resp, err := suite.Client.Get(url)
			suite.Require().NoError(err)
			suite.Require().Equal(http.StatusOK, resp.StatusCode, fmt.Sprintf("Chain %s HTTP code %d", ch.ChainName, resp.StatusCode))

			data, err := ioutil.ReadAll(resp.Body)
			suite.Require().NoError(err)

			err = resp.Body.Close()
			suite.Require().NoError(err)

			var undelegations api.UnbondingDelegationsResponse
			suite.Require().NoError(json.Unmarshal(data, &undelegations))
			suite.Require().NotEmpty(undelegations.UnbondingDelegations)

			// assert
			for _, undelegation := range undelegations.UnbondingDelegations {
				// check for same chain name
				if undelegation.ChainName == ch.ChainName {
					validatorFound := false
					for _, expected := range expectedUndelegations {
						// get hex validator address
						valAddr, err := sdk.ValAddressFromBech32(expected.ValidatorAddress)
						suite.Require().NoError(err)

						if undelegation.ValidatorAddress == hex.EncodeToString(valAddr) {
							validatorFound = true
							suite.Require().Len(undelegation.Entries, len(expected.Entries))

							// compare entries
							for _, undelegationEntry := range undelegation.Entries {
								entryFound := false
								for _, expectedEntry := range expected.Entries {
									if expectedEntry.CreationHeight == undelegationEntry.CreationHeight {
										entryFound = true
										suite.Require().Equal(expectedEntry.Balance.String(), undelegationEntry.Balance)
										suite.Require().Equal(expectedEntry.InitialBalance.String(), undelegationEntry.InitialBalance)
										entryTime, err := time.Parse(time.RFC3339, undelegationEntry.CompletionTime)
										suite.Require().NoError(err)
										suite.Require().Equal(expectedEntry.CompletionTime, entryTime)
									}
								}
								suite.Require().True(entryFound)
							}
						}
					}
					suite.Require().True(validatorFound)
				}
			}
		})
	}
}
