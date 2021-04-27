package trace

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Register(router *gin.Engine) {
	router.GET("/trace/verify/:chain/:denom", VerifyTrace)
}

func VerifyTrace(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{})
}
