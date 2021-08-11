package liquidity

import (
	"fmt"
	"net/http"

	"github.com/allinbits/demeris-backend/api/router/deps"
	"github.com/gin-gonic/gin"
)

const (
	grpcPort = 9090
)

func Register(router *gin.Engine) {
	group := router.Group("/cosmos/liquidity/v1beta1")
	group.GET("/pools", GetPools)
	group.GET("/params", GetParams)
}

// GetPools returns the of all pools.
// @Summary Gets pools info.
// @Tags pools
// @ID pools
// @Description Gets info of all pools.
// @Produce json
// @Success 200 {object} types.Pools
// @Failure 500,403 {object} deps.Error
// @Router /cosmos/liquidity/v1beta1/pools [get]
func GetPools(c *gin.Context) {
	d := deps.GetDeps(c)

	res, err := d.Store.GetPools("pools")
	if err != nil {
		e := deps.NewError(
			"pools",
			fmt.Errorf("cannot retrieve pools"),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot query pools",
			"id",
			e.ID,
			"error",
			err,
		)

		return
	}

	c.JSON(http.StatusOK, res)
}

// GetParams returns the params of liquidity module.
// @Summary Gets params of liquidity module.
// @Tags params
// @ID params
// @Description Gets params of liquidity module.
// @Produce json
// @Success 200 {object} types.Params
// @Failure 500,403 {object} deps.Error
// @Router /cosmos/liquidity/v1beta1/params [get]
func GetParams(c *gin.Context) {
	d := deps.GetDeps(c)

	res, err := d.Store.GetParams("params")
	if err != nil {
		e := deps.NewError(
			"params",
			fmt.Errorf("cannot retrieve params"),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot retrieve params",
			"id",
			e.ID,
			"error",
			err,
		)

		return
	}

	c.JSON(http.StatusOK, res)
}
