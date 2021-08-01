package chains

import (
	"fmt"
	"net/http"

	"github.com/allinbits/demeris-backend/api/router/deps"
	"github.com/gin-gonic/gin"
)

// GetFee returns the fee average in dollar for the specified chain.
// @Summary Gets average fee in dollar by chain name.
// @Tags Chain
// @ID fee
// @Description Gets average fee in dollar by chain name.
// @Param chainName path string true "chain name"
// @Produce json
// @Success 200 {object} feeResponse
// @Failure 500,403 {object} deps.Error
// @Router /chain/{chainName}/fee [get]
func GetFee(c *gin.Context) {
	var res feeResponse

	d := deps.GetDeps(c)

	chainName := c.Param("chain")

	chain, err := d.Database.Chain(chainName)

	if err != nil {
		e := deps.NewError(
			"fee",
			fmt.Errorf("cannot retrieve chain with name %v", chainName),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
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

	res = feeResponse{
		Denoms: chain.FeeTokens(),
	}

	c.JSON(http.StatusOK, res)
}

// GetFeeAddress returns the fee address for a given chain, looked up by the chain name attribute.
// @Summary Gets address to pay fee for by chain name.
// @Tags Chain
// @ID feeaddress
// @Description Gets address to pay fee for by chain name.
// @Param chainName path string true "chain name"
// @Produce json
// @Success 200 {object} feeAddressResponse
// @Failure 500,403 {object} deps.Error
// @Router /chain/{chainName}/address [get]
func GetFeeAddress(c *gin.Context) {
	var res feeAddressResponse

	d := deps.GetDeps(c)

	chainName := c.Param("chain")

	chain, err := d.Database.Chain(chainName)

	if err != nil {
		e := deps.NewError(
			"feeaddress",
			fmt.Errorf("cannot retrieve chain with name %v", chainName),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
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
		FeeAddress: chain.DemerisAddresses,
	}

	c.JSON(http.StatusOK, res)
}

// GetFeeAddresses returns the fee address for all chains.
// @Summary Gets all addresses to pay fee for.
// @Tags Chain
// @ID feeaddresses
// @Description Gets all addresses to pay fee for.
// @Produce json
// @Success 200 {object} feeAddressesResponse
// @Failure 500,403 {object} deps.Error
// @Router /chains/fee/addresses [get]
func GetFeeAddresses(c *gin.Context) {
	var res feeAddressesResponse

	d := deps.GetDeps(c)

	chains, err := d.Database.Chains()

	if err != nil {
		e := deps.NewError(
			"feeaddress",
			fmt.Errorf("cannot retrieve chains"),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
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
			feeAddress{
				ChainName:  c.ChainName,
				FeeAddress: c.DemerisAddresses,
			},
		)
	}

	c.JSON(http.StatusOK, res)
}

// GetFeeToken returns the fee token for a given chain, looked up by the chain name attribute.
// @Summary Gets token used to pay fees by chain name.
// @Tags Chain
// @ID feetoken
// @Description Gets token used to pay fees by chain name.
// @Param chainName path string true "chain name"
// @Produce json
// @Success 200 {object} feeTokenResponse
// @Failure 500,403 {object} deps.Error
// @Router /chain/{chainName}/token [get]
func GetFeeToken(c *gin.Context) {
	var res feeTokenResponse

	d := deps.GetDeps(c)

	chainName := c.Param("chain")

	chain, err := d.Database.Chain(chainName)

	if err != nil {
		e := deps.NewError(
			"feetoken",
			fmt.Errorf("cannot retrieve chain with name %v", chainName),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
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

	for _, cc := range chain.FeeTokens() {
		res.FeeTokens = append(res.FeeTokens, cc)
	}

	c.JSON(http.StatusOK, res)
}
