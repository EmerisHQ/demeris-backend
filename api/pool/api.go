package pool

import (
	"github.com/gin-gonic/gin"
)

func Register(router *gin.Engine) {
	rel := router.Group("/pool")

	rel.GET("pool/:poolId")

}

// GetSwapFee returns the swap fee of past 1 hour n.
// @Summary Gets ticket by id.
// @Tags pool
// @ID swap fee
// @Description Gets swap fee of past one hour by pool id.
// @Param pool path string true "pool id"
// @Produce json
// @Success 200 {object} sdk.Coins
// @Failure 500,403 {object} deps.Error
// @Router /pool/{poolID}/swapfee [get]
func GetSwapFee(c *gin.Context) {
	//
	//d := deps.GetDeps(c)
	//
	//poolId := c.Param("poolId")
}
