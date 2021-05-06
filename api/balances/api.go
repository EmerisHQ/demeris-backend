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

// GetBalancesByAddress returns balances of an address.
// @Summary Gets address balance
// @Tags Balances
// @ID get-balances
// @Description gets address balance
// @Produce json
// @Param address path string true "address to query staking balances for"
// @Success 200 {object} balancesResponse
// @Failure 500,403 {object} deps.Error
// @Router /balances/{address} [get]
func GetBalancesByAddress(c *gin.Context) {

	var res balancesResponse
	d := deps.GetDeps(c)

	address := c.Param("address")

	d.Logger.Info("Searching for addresses, ", address)

	balances, err := d.Database.Balances(address)

	if err != nil {
		e := deps.NewError(
			"balances",
			fmt.Errorf("cannot retrieve balances for address %v", address),
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
					"balances",
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
