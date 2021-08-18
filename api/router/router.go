package router

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/allinbits/demeris-backend/api/block"

	"github.com/allinbits/demeris-backend/utils/logging"

	"github.com/allinbits/demeris-backend/api/relayer"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/allinbits/demeris-backend/utils/validation"
	"github.com/gin-contrib/cache/persistence"
	"github.com/gin-gonic/gin/binding"

	kube "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/allinbits/demeris-backend/api/chains"
	"github.com/allinbits/demeris-backend/api/tx"
	"github.com/allinbits/demeris-backend/api/verifieddenoms"

	"github.com/allinbits/demeris-backend/api/account"
	"github.com/allinbits/demeris-backend/api/database"
	"github.com/allinbits/demeris-backend/api/router/deps"
	"github.com/allinbits/demeris-backend/utils/store"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Router struct {
	g            *gin.Engine
	db           *database.Database
	l            *zap.SugaredLogger
	s            *store.Store
	k8s          kube.Client
	k8sNamespace string
	cdc          codec.Marshaler
	cnsURL       string
	cache        *persistence.InMemoryStore
}

func New(
	db *database.Database,
	l *zap.SugaredLogger,
	s *store.Store,
	kubeClient kube.Client,
	kubeNamespace string,
	cnsURL string,
	cdc codec.Marshaler,
	debug bool,
) *Router {
	gin.SetMode(gin.ReleaseMode)

	if debug {
		gin.SetMode(gin.DebugMode)
	}

	engine := gin.New()

	store := persistence.NewInMemoryStore(10 * time.Second)

	r := &Router{
		g:            engine,
		db:           db,
		l:            l,
		s:            s,
		k8s:          kubeClient,
		k8sNamespace: kubeNamespace,
		cnsURL:       cnsURL,
		cdc:          cdc,
		cache:        store,
	}

	r.metrics()

	validation.JSONFields(binding.Validator)

	engine.Use(r.catchPanics())
	if debug {
		engine.Use(logging.LogRequest(l.Desugar()))
	}
	engine.Use(r.decorateCtxWithDeps())
	engine.Use(r.handleErrors())
	engine.RedirectTrailingSlash = false
	engine.RedirectFixedPath = false

	registerRoutes(engine, store)

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
			Logger:        r.l,
			Database:      r.db,
			Store:         r.s,
			CNSURL:        r.cnsURL,
			KubeNamespace: r.k8sNamespace,
			Codec:         r.cdc,
			K8S:           &r.k8s,
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

func registerRoutes(engine *gin.Engine, store *persistence.InMemoryStore) {
	// @tag.name Account
	// @tag.description Account-querying endpoints
	account.Register(engine)

	// @tag.name Denoms
	// @tag.description Denoms-related endpoints
	verifieddenoms.Register(engine, store)

	// @tag.name Chain
	// @tag.description Chain-related endpoints
	chains.Register(engine, store)

	// @tag.name Transactions
	// @tag.description Transaction-related endpoints
	tx.Register(engine)

	// @tag.name Relayer
	// @tag.description Relayer-related endpoints
	relayer.Register(engine, store)

	// @tag.name Block
	// @tag.description Blocks-related endpoints
	block.Register(engine)
}
