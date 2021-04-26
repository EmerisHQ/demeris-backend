package router

import (
	"net/http"

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
	engine.Use(r.handleErrors())

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

func (r *Router) handleErrors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) != 0 {
			c.JSON(http.StatusBadRequest, c.Errors.JSON())
		}
	}
}

func registerRoutes(engine *gin.Engine) {
	balances.Register(engine)
	trace.Register(engine)
}
