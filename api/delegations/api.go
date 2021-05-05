package delegations

import (
	"fmt"
	"net/http"

	"github.com/allinbits/demeris-backend/api/router/deps"
	"github.com/gin-gonic/gin"
)

func Register(router *gin.Engine) {
	router.GET("/staking_balances/:address", GetDelegationsByAddress)
}

// GetDelegationsByAddress returns staking balances of an address.
// @Summary Gets staking balance
// @Description gets staking balance
// @Produce json
// Param address query string true "staking balance search by q"
// @Success 200 {object} delegations.Delegations
// @Failure 500,403 {object} deps.Error
// @Router /staking_balances [get]
func GetDelegationsByAddress(c *gin.Context) {
	var res Delegations

	d, err := deps.GetDeps(c)
	if err != nil {
		c.Error(deps.NewError(
			"delegations",
			fmt.Errorf("internal error"),
			http.StatusInternalServerError,
		))

		panic("cannot retrieve context deps")
		return
	}

	address := c.Param("address")

	dl, err := d.Database.Delegations(address)

	if err != nil {
		e := deps.NewError(
			"delegations",
			fmt.Errorf("cannot retrieve delegations for address %v", address),
			http.StatusBadRequest,
		)

		c.Error(e)

		d.Logger.Errorw(
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
		res.Delegations = append(res.Delegations, Delegation{
			ValidatorAddress: del.Validator,
			Amount:           del.Amount,
			ChainName:        del.ChainName,
		})
	}

	res.Delegator = address

	c.JSON(http.StatusOK, res)
}
