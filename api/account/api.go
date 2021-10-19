package account

import (
	"fmt"
	"net/http"

	"github.com/allinbits/demeris-backend/api/database"
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
	group.GET("/tickets", GetUserTickets)
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

	vd, err := verifiedDenomsMap(d.Database)
	if err != nil {
		e := deps.NewError(
			"account",
			fmt.Errorf("cannot retrieve account for address %v", address),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot query database verified denoms",
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
			Address: b.Address,
			Amount:  b.Amount,
			OnChain: b.ChainName,
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
			balance.Verified = vd[denomTrace.BaseDenom]
		} else {
			balance.Verified = vd[b.Denom]
			balance.BaseDenom = b.Denom
		}

		res.Balances = append(res.Balances, balance)
	}

	c.JSON(http.StatusOK, res)
}

func verifiedDenomsMap(d *database.Database) (map[string]bool, error) {
	chains, err := d.VerifiedDenoms()
	if err != nil {
		return nil, err
	}

	ret := make(map[string]bool)
	for _, cc := range chains {
		for _, vd := range cc {
			ret[vd.Name] = vd.Verified
		}
	}

	return ret, err
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
// @ID get-all-numbers-account
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

	numbers, err := d.Database.Numbers(address)

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
	}

	res.Numbers = numbers

	c.JSON(http.StatusOK, res)
}

func GetUserTickets(c *gin.Context) {
	d := deps.GetDeps(c)

	address := c.Param("address")

	tickets, err := d.Store.GetUserTickets(address)
	if err != nil {
		e := deps.NewError(
			"tickets",
			fmt.Errorf("cannot retrieve tickets for address %v", address),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot query store for tickets",
			"address",
			address,
			"error",
			err,
		)

		return
	}

	c.JSON(http.StatusOK, userTicketsResponse{Tickets: tickets})
}
