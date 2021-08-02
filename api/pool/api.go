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

// GetSwapFee returns the swap fee of past 1 hour n.
// @Summary Gets swap fee by id.
// @Tags pool
// @ID swap fee
// @Description Gets swap fee of past one hour by pool id.
// @Param pool path string true "pool id"
// @Produce json
// @Success 200 {object} sdk.Coins
// @Failure 500,403 {object} deps.Error
// @Router /pool/{poolID}/swapfee [get]
func getSwapFee(c *gin.Context) {

	d := deps.GetDeps(c)

	poolId := c.Param("poolId")

	coins, err := d.Store.GetSwapFees(poolId)
	if err != nil {
		e := deps.NewError(
			"status",
			fmt.Errorf("cannot retrieve relayer status"),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot retrieve relayer status",
			"id",
			e.ID,
			"error",
			err,
		)

		return
	}

	c.JSON(http.StatusOK, SwapFeesResponse{Coins: coins})
}
