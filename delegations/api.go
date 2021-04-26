package delegations

import (
	"fmt"
	"net/http"

	"github.com/allinbits/navigator-backend/router/deps"
	"github.com/gin-gonic/gin"
)

func Register(router *gin.Engine) {
	router.GET("/staking_balances/:address", GetDelegationsByAddress)
}

// GetDelegationsByAddress returns staking balances of an address.
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
			ValidatorAddress: del.ValidatorAddress,
			Amount:           del.Amount,
		})
	}

	res.Delegator = address

	c.JSON(http.StatusOK, res)
}
