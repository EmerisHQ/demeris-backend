package feeaddress

import (
	"fmt"
	"net/http"

	"github.com/allinbits/demeris-backend/api/router/deps"
	"github.com/gin-gonic/gin"
)

func Register(router *gin.Engine) {
	router.GET("/fee_address", GetFeeAddresses)
	router.GET("/fee_address/:chain", GetFeeAddress)
}

// GetFeeAddresses returns the fee address for all chains.
func GetFeeAddresses(c *gin.Context) {
	var res feeAddressesResponse

	d, err := deps.GetDeps(c)
	if err != nil {
		c.Error(deps.NewError(
			"feeaddress",
			fmt.Errorf("internal error"),
			http.StatusInternalServerError,
		))

		panic("cannot retrieve context deps")
		return
	}

	chains, err := d.Database.Chains()

	if err != nil {
		e := deps.NewError(
			"feeaddress",
			fmt.Errorf("cannot retrieve chains"),
			http.StatusBadRequest,
		)

		c.Error(e)

		d.Logger.Errorw(
			"cannot retrieve chains",
			"id",
			e.ID,
			"error",
			err,
		)

		return
	}

	for _, c := range chains {
		res.FeeAddresses = append(
			res.FeeAddresses,
			feeAddressResponse{
				ChainName:  c.ChainName,
				FeeAddress: c.FeeAddress,
			},
		)
	}

	c.JSON(http.StatusOK, res)
}

// GetFeeAddress returns the fee address for a given chain, looked up by the chain name attribute.
func GetFeeAddress(c *gin.Context) {
	var res feeAddressResponse

	d, err := deps.GetDeps(c)
	if err != nil {
		c.Error(deps.NewError(
			"feeaddress",
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
			"feeaddress",
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

	res = feeAddressResponse{
		ChainName:  chain.ChainName,
		FeeAddress: chain.FeeAddress,
	}

	c.JSON(http.StatusOK, res)
}
