package feetoken

import (
	"fmt"
	"net/http"

	"github.com/allinbits/demeris-backend/api/router/deps"
	"github.com/gin-gonic/gin"
)

func Register(router *gin.Engine) {
	router.GET("/fee_token/:chain", GetFeeToken)
}

// GetFeeToken returns the fee token for a given chain, looked up by the chain name attribute.
func GetFeeToken(c *gin.Context) {
	var res []FeeToken

	d, err := deps.GetDeps(c)
	if err != nil {
		c.Error(deps.NewError(
			"feetoken",
			fmt.Errorf("internal error"),
			http.StatusInternalServerError,
		))

		panic("cannot retrieve context deps")
		return
	}

	chainName := c.Param("chain")

	chain, err := d.Database.Chain(chainName)

	if err != nil {
		e := deps.NewError(
			"feetoken",
			fmt.Errorf("cannot retrieve chain with name %v", chainName),
			http.StatusBadRequest,
		)

		c.Error(e)

		d.Logger.Errorw(
			"cannot retrieve chain",
			"id",
			e.ID,
			"name",
			chainName,
			"error",
			err,
		)

		return
	}

	for _, cc := range chain.FeeTokens {
		res = append(res, FeeToken{
			Name:     cc.Name,
			Verified: cc.IsVerified,
		})
	}

	c.JSON(http.StatusOK, res)
}
