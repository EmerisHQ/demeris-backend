package router

import (
	"github.com/allinbits/navigator-backend/balances"
	"github.com/allinbits/navigator-backend/database"
	"github.com/allinbits/navigator-backend/trace"
	"github.com/allinbits/navigator-utils/logging"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Router struct {
	g  *gin.Engine
	db *database.Database
	l  *zap.SugaredLogger
}

func New(db *database.Database, l *zap.SugaredLogger) *Router {
	router := gin.Default()

	router.Use(logging.LogRequest(l.Desugar()))

	registerRoutes(router)

	return &Router{
		g:  router,
		db: db,
		l:  l,
	}
}

func (r *Router) Serve(address string) error {
	return r.g.Run(address)
}

func registerRoutes(engine *gin.Engine) {
	balances.Register(engine)
	trace.Register(engine)
}
