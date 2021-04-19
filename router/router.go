package router

import (
	"github.com/allinbits/navigator-backend/balances"
	"github.com/allinbits/navigator-backend/database"
	"github.com/allinbits/navigator-backend/router/deps"
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
	engine := gin.Default()

	r := &Router{
		g:  engine,
		db: db,
		l:  l,
	}

	engine.Use(logging.LogRequest(l.Desugar()))
	engine.Use(r.decorateCtxWithDeps())

	registerRoutes(engine)

	return r
}

func (r *Router) Serve(address string) error {
	return r.g.Run(address)
}

func (r *Router) decorateCtxWithDeps() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("deps", &deps.Deps{
			Logger:   r.l,
			Database: r.db,
		})
	}
}

func registerRoutes(engine *gin.Engine) {
	balances.Register(engine)
	trace.Register(engine)
}
