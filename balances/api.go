package balances

import (
	"net/http"
	"strings"

	"github.com/allinbits/navigator-backend/router/deps"
	"github.com/gin-gonic/gin"
)

func Register(router *gin.Engine) {
	router.GET("/balances/:addresses", GetBalancesByAddresses)
}

// GetBalancesByAddresses - Find balances by addresses
func GetBalancesByAddresses(c *gin.Context) {

	res := []Balance{}
	d, err := deps.GetDeps(c)
	if err != nil {
		c.Error(err)
		return
	}

	addresses := strings.Split(c.Param("addresses"), ",")

	d.Logger.Info("Searching for addresses, ", addresses)

	balances, err := d.Database.Balances(addresses)

	if err != nil {
		d.Logger.Errorw("cannot query database balance for addresses", "addresses", addresses, "error", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	// TODO: get unique chains
	// perhaps we can remove this since there will be another endpoint specifically for fee tokens

	for _, b := range balances {
		balance := Balance{
			Address:  b.Address,
			Amount:   b.Amount,
			Verified: true,
			OnChain:  b.ChainName,
		}

		if b.Denom[:4] == "ibc/" {
			// is ibc token
			balance.Ibc = IbcInfo{
				IbcDenom: b.Denom,
			}
			balance.Native = false

			// TODO: verify trace
			balance.BaseDenom = "" // check trace

		} else {
			balance.Native = true
			balance.Verified = true
			balance.BaseDenom = b.Denom
		}

		res = append(res, balance)
	}
	// d.Logger.Info(d.Database.Balances(addresses))
	d.Logger.Info(balances)

	c.JSON(http.StatusOK, gin.H{
		"balances": res,
	})
}
