package router

import (
	"errors"

	"github.com/allinbits/demeris-backend/api/chains"
	"github.com/allinbits/demeris-backend/api/feetoken"
	"github.com/allinbits/demeris-backend/api/verifieddenoms"

	"github.com/allinbits/demeris-backend/api/delegations"

	"github.com/allinbits/demeris-backend/api/balances"
	"github.com/allinbits/demeris-backend/api/database"
	"github.com/allinbits/demeris-backend/api/denom"
	"github.com/allinbits/demeris-backend/api/router/deps"
	"github.com/allinbits/demeris-backend/utils/logging"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Router struct {
	g      *gin.Engine
	db     *database.Database
	l      *zap.SugaredLogger
	cnsURL string
}

func New(db *database.Database, l *zap.SugaredLogger, cnsURL string) *Router {
	engine := gin.Default()

	r := &Router{
		g:      engine,
		db:     db,
		l:      l,
		cnsURL: cnsURL,
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
			CNSURL:   r.cnsURL,
		})
	}
}

func (r *Router) handleErrors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		l := c.Errors.Last()
		if l == nil {
			c.Next()
			return
		}

		rerr := deps.Error{}
		if !errors.As(l, &rerr) {
			panic(l)
		}

		c.JSON(rerr.StatusCode, rerr)
	}
}

func registerRoutes(engine *gin.Engine) {
	balances.Register(engine)
	denom.Register(engine)
	delegations.Register(engine)
	feetoken.Register(engine)
	verifieddenoms.Register(engine)
	chains.Register(engine)
}
