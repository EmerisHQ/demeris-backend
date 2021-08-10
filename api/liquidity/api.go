package liquidity

import (
	"github.com/allinbits/demeris-backend/api/router/deps"
	"github.com/gin-gonic/gin"
)

const (
	grpcPort = 9090
)

func Register(router *gin.Engine) {
	group := router.Group("/cosmos/liquidity")
	group.GET("/pools", GetPools)
}

func GetPools(c *gin.Context) {
	d := deps.GetDeps(c)

	address := c.Param("address")

	pools, err := d.Database.Pools()
}
