package router

import (
	"github.com/allinbits/navigator-backend/balances"
	"github.com/allinbits/navigator-backend/trace"
	"github.com/gin-gonic/gin"
)

func New() *gin.Engine {
	router := gin.Default()
	registerRoutes(router)
	return router
}

func registerRoutes(engine *gin.Engine) {
	balances.Register(engine)
	trace.Register(engine)
}
