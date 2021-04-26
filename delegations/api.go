package delegations

import (
	"fmt"
	"net/http"

	"github.com/allinbits/navigator-backend/router/deps"
	"github.com/gin-gonic/gin"
)

func Register(router *gin.Engine) {
	router.GET("/delegations/:address", GetDelegationsByAddress)
}

// GetStakingBalanceByAddress - Find balances by addresses
func GetDelegationsByAddress(c *gin.Context) {
	var res Delegations

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

	dl, err := d.Database.Delegations(address)

	if err != nil {
		c.Error(deps.NewError(
			"balances",
			fmt.Errorf("cannot retrieve delegations for address %v", address),
			http.StatusBadRequest,
		))

		d.Logger.Errorw(
			"cannot query database delegations for addresses",
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
