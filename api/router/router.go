package router

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/allinbits/demeris-backend/api/chains"
	"github.com/allinbits/demeris-backend/api/tx"
	"github.com/allinbits/demeris-backend/api/verifieddenoms"

	"github.com/allinbits/demeris-backend/api/delegations"

	"github.com/allinbits/demeris-backend/api/balances"
	"github.com/allinbits/demeris-backend/api/database"
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

	r.metrics()

	engine.Use(r.catchPanics())
	engine.Use(logging.LogRequest(l.Desugar()))
	engine.Use(r.decorateCtxWithDeps())
	engine.Use(r.handleErrors())
	engine.RedirectTrailingSlash = false
	engine.RedirectFixedPath = false

	registerRoutes(engine)

	return r
}

func (r *Router) Serve(address string) error {
	return r.g.Run(address)
}

func (r *Router) catchPanics() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rval := recover(); rval != nil {
				// okay we panic-ed, log it through r's logger and write back internal server error
				err := deps.NewError(
					"fatal_error",
					errors.New("internal server error"),
					http.StatusInternalServerError)

				r.l.Errorw(
					"panic handler triggered while handling call",
					"endpoint", c.Request.RequestURI,
					"error", fmt.Sprint(rval),
					"error_id", err.ID,
				)

				c.AbortWithStatusJSON(
					http.StatusInternalServerError,
					err,
				)

				return
			}
		}()
		c.Next()
	}
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
	// @tag.name Balances
	// @tag.description Balances-querying endpoints
	balances.Register(engine)
	delegations.Register(engine)

	// @tag.name Denoms
	// @tag.description Denoms-related endpoints
	verifieddenoms.Register(engine)

	// @tag.name Chain
	// @tag.description Chain-related endpoints
	chains.Register(engine)

	// @tag.name Transactions
	// @tag.description Transaction-related endpoints
	tx.Register(engine)
}
