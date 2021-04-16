package balances

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

func Register(router *gin.Engine) {
	router.GET("/balances/:addresses", GetBalancesByAddresses)
}
// GetBalancesByAddresses - Find balances by addresses
func GetBalancesByAddresses(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{})
}
