package balances

import (
	"net/http"

	"github.com/allinbits/navigator-backend/router/deps"
	"github.com/gin-gonic/gin"
)

func Register(router *gin.Engine) {
	router.GET("/balances/:addresses", GetBalancesByAddresses)
}

// GetBalancesByAddresses - Find balances by addresses
func GetBalancesByAddresses(c *gin.Context) {
	d, err := deps.GetDeps(c)
	if err != nil {
		c.Error(err)
		return
	}

	d.Logger.Info("deps works!")
	c.JSON(http.StatusOK, gin.H{})
}
