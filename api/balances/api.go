package balances

import (
	"fmt"
	"net/http"

	"github.com/allinbits/demeris-backend/api/router/deps"
	"github.com/gin-gonic/gin"
)

func Register(router *gin.Engine) {
	router.GET("/balances/:address", GetBalancesByAddress)
}

// GetBalancesByAddress - Find balances by address
func GetBalancesByAddress(c *gin.Context) {

	res := []Balance{}
	d, err := deps.GetDeps(c)
	if err != nil {
		c.Error(deps.NewError(
			"balances",
			fmt.Errorf("internal error"),
			http.StatusInternalServerError,
		))

		panic("cannot retrieve context deps")
		return
	}

	address := c.Param("address")

	d.Logger.Info("Searching for addresses, ", address)

	balances, err := d.Database.Balances(address)

	if err != nil {
		e := deps.NewError(
			"balances",
			fmt.Errorf("cannot retrieve balances for address %v", address),
			http.StatusBadRequest,
		)

		c.Error(e)

		d.Logger.Errorw(
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
		balance := Balance{
			Address:  b.Address,
			Amount:   b.Amount,
			Verified: true,
			OnChain:  b.ChainName,
		}

		if b.Denom[:4] == "ibc/" {
			// is ibc token
			balance.Ibc = IbcInfo{
				Hash: b.Denom[4:],
			}

			denomTrace, err := d.Database.DenomTrace(b.ChainName, b.Denom[4:])

			if err != nil {
				e := deps.NewError(
					"balances",
					fmt.Errorf("cannot query denom trace for token %v on chain %v", b.Denom, b.ChainName),
					http.StatusBadRequest,
				)

				c.Error(e)

				d.Logger.Errorw(
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

		res = append(res, balance)
	}
	// d.Logger.Info(d.Database.Balances(addresses))
	d.Logger.Info(balances)

	c.JSON(http.StatusOK, gin.H{
		"balances": res,
	})
}
