package pool

import (
	"fmt"
	"net/http"

	"github.com/allinbits/demeris-backend/api/router/deps"
	"github.com/gin-gonic/gin"
)

func Register(router *gin.Engine) {
	rel := router.Group("/pool/:poolId")

	rel.GET("/swapfees", getSwapFee)

}

// getSwapFee returns the swap fee of past 1 hour n.
// @Summary Gets swap fee by pool id.
// @Tags pool
// @ID swap fee
// @Description Gets swap fee of past one hour by pool id.
// @Param pool path string true "pool id"
// @Produce json
// @Success 200 {object} types.Coins
// @Failure 500,403 {object} deps.Error
// @Router /pool/{poolID}/swapfees [get]
func getSwapFee(c *gin.Context) {

	d := deps.GetDeps(c)

	poolId := c.Param("poolId")

	fees, err := d.Store.GetSwapFees(poolId)
	if err != nil {
		e := deps.NewError(
			"swap fees",
			fmt.Errorf("cannot get swap fees"),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot get swap fees",
			"id",
			e.ID,
			"poolId",
			poolId,
			"error",
			err,
		)

		return
	}

	c.JSON(http.StatusOK, SwapFeesResponse{Fees: fees})
}
