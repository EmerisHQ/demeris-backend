package cached

import (
	"fmt"
	"net/http"

	"github.com/allinbits/demeris-backend/api/router/deps"
	"github.com/gin-gonic/gin"
)

func Register(router *gin.Engine) {
	group := router.Group("/cached/cosmos/v1beta1")

	group.GET("/pools", getPools)
	group.GET("/params", getParams)
	group.GET("/supply", getSupply)
}

// getPools returns the of all pools.
// @Summary Gets pools info.
// @Tags pools
// @ID pools
// @Description Gets info of all pools.`10
// @Produce json
// @Success 200 {object} types.Pools
// @Failure 500,403 {object} deps.Error
// @Router /cosmos/liquidity/v1beta1/pools [get]
func getPools(c *gin.Context) {
	d := deps.GetDeps(c)

	res, err := d.Store.GetPools()
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

// getParams returns the params of liquidity module.
// @Summary Gets params of liquidity module.
// @Tags params
// @ID params
// @Description Gets params of liquidity module.
// @Produce json
// @Success 200 {object} types.Params
// @Failure 500,403 {object} deps.Error
// @Router /cosmos/liquidity/v1beta1/params [get]
func getParams(c *gin.Context) {
	d := deps.GetDeps(c)

	res, err := d.Store.GetParams()
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

// getSupply returns the total supply.
// @Summary Gets total supply of cosmos-hub
// @Tags supply
// @ID supply
// @Description Gets total supply of cosmos hub.
// @Produce json
// @Success 200 {object} types.Coins
// @Failure 500,403 {object} deps.Error
// @Router / [get]
func getSupply(c *gin.Context) {
	d := deps.GetDeps(c)

	res, err := d.Store.GetSupply()
	if err != nil {
		e := deps.NewError(
			"supply",
			fmt.Errorf("cannot retrieve total supply"),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot retrieve total supply",
			"id",
			e.ID,
			"error",
			err,
		)

		return
	}

	c.JSON(http.StatusOK, res)
}
