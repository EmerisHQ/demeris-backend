package account

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/allinbits/demeris-backend/api/database"
	"github.com/allinbits/demeris-backend/models"
	"github.com/cosmos/cosmos-sdk/simapp"
	bech322 "github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"

	"google.golang.org/grpc"
	"net/http"

	"github.com/allinbits/demeris-backend/api/router/deps"
	"github.com/gin-gonic/gin"
)

const (
	grpcPort = 9090
)

func Register(router *gin.Engine) {
	group := router.Group("/account/:address")
	group.GET("/balance", GetBalancesByAddress)
	group.GET("/stakingbalances", GetDelegationsByAddress)
	group.GET("/numbers", GetNumbersByAddress)
}

// GetBalancesByAddress returns account of an address.
// @Summary Gets address balance
// @Tags Account
// @ID get-account
// @Description gets address balance
// @Produce json
// @Param address path string true "address to query balance for"
// @Success 200 {object} balancesResponse
// @Failure 500,403 {object} deps.Error
// @Router /account/{address}/balance [get]
func GetBalancesByAddress(c *gin.Context) {
	var res balancesResponse
	d := deps.GetDeps(c)

	address := c.Param("address")

	d.Logger.Info("Searching for addresses, ", address)

	balances, err := d.Database.Balances(address)

	if err != nil {
		e := deps.NewError(
			"account",
			fmt.Errorf("cannot retrieve account for address %v", address),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot query database balance for address",
			"id",
			e.ID,
			"address",
			address,
			"error",
			err,
		)
		return
	}

	// TODO: get unique chains
	// perhaps we can remove this since there will be another endpoint specifically for fee tokens

	for _, b := range balances {
		balance := balance{
			Address:  b.Address,
			Amount:   b.Amount,
			Verified: true,
			OnChain:  b.ChainName,
		}

		if b.Denom[:4] == "ibc/" {
			// is ibc token
			balance.Ibc = ibcInfo{
				Hash: b.Denom[4:],
			}

			denomTrace, err := d.Database.DenomTrace(b.ChainName, b.Denom[4:])

			if err != nil {
				e := deps.NewError(
					"account",
					fmt.Errorf("cannot query denom trace for token %v on chain %v", b.Denom, b.ChainName),
					http.StatusBadRequest,
				)

				d.WriteError(c, e,
					"cannot query database balance for address",
					"id",
					e.ID,
					"token",
					b.Denom,
					"chain",
					b.ChainName,
					"error",
					err,
				)

				return
			}
			balance.BaseDenom = denomTrace.BaseDenom
			balance.Ibc.Path = denomTrace.Path

		} else {
			balance.Verified = true
			balance.BaseDenom = b.Denom
		}

		res.Balances = append(res.Balances, balance)
	}
	// d.Logger.Info(d.Database.Balances(addresses))
	d.Logger.Info(balances)

	c.JSON(http.StatusOK, res)
}

// GetDelegationsByAddress returns staking account of an address.
// @Summary Gets staking balance
// @Description gets staking balance
// @Tags Account
// @ID get-staking-account
// @Produce json
// @Param address path string true "address to query staking for"
// @Success 200 {object} stakingBalancesResponse
// @Failure 500,403 {object} deps.Error
// @Router /account/{address}/stakingbalance [get]
func GetDelegationsByAddress(c *gin.Context) {
	var res stakingBalancesResponse

	d := deps.GetDeps(c)

	address := c.Param("address")

	dl, err := d.Database.Delegations(address)

	if err != nil {
		e := deps.NewError(
			"delegations",
			fmt.Errorf("cannot retrieve delegations for address %v", address),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot query database delegations for addresses",
			"id",
			e.ID,
			"address",
			address,
			"error",
			err,
		)

		return
	}

	for _, del := range dl {
		res.StakingBalances = append(res.StakingBalances, stakingBalance{
			ValidatorAddress: del.Validator,
			Amount:           del.Amount,
			ChainName:        del.ChainName,
		})
	}

	c.JSON(http.StatusOK, res)
}

// GetNumbersByAddress returns sequence and account number of an address.
// @Summary Gets sequence and account number
// @Description Gets sequence and account number
// @Tags Account
// @ID get-numbers-account
// @Produce json
// @Param address path string true "address to query numbers for"
// @Success 200 {object} numbersResponse
// @Failure 500,403 {object} deps.Error
// @Router /account/{address}/numbers [get]
func GetNumbersByAddress(c *gin.Context) {
	var res numbersResponse

	d := deps.GetDeps(c)

	address := c.Param("address")

	dd, err := d.Database.ChainNames()
	d.Logger.Debugw("chain names", "chain names", dd, "error", err)

	/*
		PSA: do not remove this comment, this is the proper tracelistener-based implementation of this endpoint,
		which will  be used some time in the future as soon as we fix the auth mismatch error.

		dl, err := d.Database.Numbers(address)

		if err != nil {
			e := deps.NewError(
				"numbers",
				fmt.Errorf("cannot retrieve account/sequence numbers for address %v", address),
				http.StatusBadRequest,
			)

			d.WriteError(c, e,
				"cannot query database auth for addresses",
				"id",
				e.ID,
				"address",
				address,
				"error",
				err,
			)

			return
		}*/

	resp, err := fetchNumbers(dd, address)
	if err != nil {
		e := deps.NewError(
			"numbers",
			fmt.Errorf("cannot retrieve account/sequence numbers for address %v", address),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot query nodes auth for addresses",
			"id",
			e.ID,
			"address",
			address,
			"error",
			err,
		)

		return
	}

	res.Numbers = resp

	c.JSON(http.StatusOK, res)
}

func fetchNumbers(cns []database.ChainName, account string) ([]models.AuthRow, error) {
	accBytes, err := hex.DecodeString(account)
	if err != nil {
		return nil, fmt.Errorf("cannot decode hex bytes from account string")
	}

	queryGroup, _ := errgroup.WithContext(context.Background())

	results := make([]models.AuthRow, len(cns))

	cdc, _ := simapp.MakeCodecs()

	for i, chain := range cns {
		addr, err := bech322.ConvertAndEncode(chain.AccountPrefix, accBytes)
		if err != nil {
			return nil, fmt.Errorf("cannot encode bytes to %s acc address, %w", chain.ChainName, err)
		}

		i, chain, addr := i, chain, addr

		queryGroup.Go(func() error {
			resp, err := queryChainNumbers(chain.ChainName, addr)
			if err != nil {
				return err
			}

			if resp == nil {
				return nil // account doesn't have numbers
			}

			// get a baseAccount
			var accountI types.AccountI

			if err := cdc.UnpackAny(resp.Account, &accountI); err != nil {
				return err
			}

			results[i] = models.AuthRow{
				TracelistenerDatabaseRow: models.TracelistenerDatabaseRow{
					ChainName: chain.ChainName,
				},
				Address:        account,
				SequenceNumber: accountI.GetSequence(),
				AccountNumber:  accountI.GetAccountNumber(),
			}

			return nil
		})
	}

	if err := queryGroup.Wait(); err != nil {
		return nil, fmt.Errorf("cannot query chains, %w", err)
	}

	return results, nil
}

func queryChainNumbers(chainName string, address string) (*types.QueryAccountResponse, error) {
	grpcConn, err := grpc.Dial(fmt.Sprintf("%s:%d", chainName, grpcPort), grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	authQuery := types.NewQueryClient(grpcConn)

	nums, err := authQuery.Account(context.Background(), &types.QueryAccountRequest{
		Address: address,
	})

	if status.Code(err) == codes.NotFound ||
		strings.Contains(strings.ToLower(err.Error()), "not found") {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("cannot query account, %w", err)
	}

	_ = grpcConn.Close()

	return nums, nil
}
