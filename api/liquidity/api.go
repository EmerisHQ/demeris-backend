package liquidity

import (
	"fmt"
	"net/http"

	"github.com/allinbits/demeris-backend/api/router/deps"
	"github.com/gin-gonic/gin"
	_ "github.com/gravity-devs/liquidity/x/liquidity/types"
)

const (
	grpcPort = 9090
)

func Register(router *gin.Engine) {
	group := router.Group("/cached/cosmos/v1beta1")

	group.GET("/swapfees", getSwapFee)
	group.GET("/pools", GetPools)
	group.GET("/params", GetParams)
	group.GET("/supply", getSupply)
}

// GetPools returns the of all pools.
// @Summary Gets pools info.
// @Tags pools
// @ID pools
// @Description Gets info of all pools.`10
// @Produce json
// @Success 200 {object} types.Pools
// @Failure 500,403 {object} deps.Error
// @Router /cosmos/liquidity/v1beta1/pools [get]
func GetPools(c *gin.Context) {
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
