package rest

import (
	"errors"
	"net/http"

	"github.com/allinbits/demeris-backend/utils/validation"
	"github.com/gin-gonic/gin/binding"

	"github.com/allinbits/demeris-backend/cns/chainwatch"

	kube "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-playground/validator/v10"

	"github.com/allinbits/demeris-backend/cns/database"
	"github.com/allinbits/demeris-backend/utils/logging"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Server struct {
	l  *zap.SugaredLogger
	gl *zap.Logger
	d  *database.Instance
	g  *gin.Engine
	k  *kube.Client
	rc *chainwatch.Connection
}

type router struct {
	s *Server
}

func NewServer(l *zap.SugaredLogger, d *database.Instance, kube *kube.Client, rc *chainwatch.Connection, debug bool) *Server {
	if !debug {
		gin.SetMode(gin.ReleaseMode)
	}

	g := gin.New()

	s := &Server{
		l:  l,
		d:  d,
		g:  g,
		k:  kube,
		rc: rc,
	}

	r := &router{s: s}

	validation.JSONFields(binding.Validator)

	g.Use(logging.LogRequest(l.Desugar()))
	g.Use(ginzap.RecoveryWithZap(l.Desugar(), true))

	g.Use(func(c *gin.Context) {

		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST,HEAD,PATCH, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	g.GET(r.getChain())
	g.GET(r.getChains())
	g.GET(r.denomsData())
	g.POST(r.addChain())
	g.POST(r.updatePrimaryChannel())
	g.POST(r.updateDenoms())
	g.DELETE(r.deleteChain())

	g.NoRoute(func(context *gin.Context) {
		e(context, http.StatusNotFound, errors.New("not found"))
	})

	return s
}

func (s *Server) Serve(where string) error {
	return s.g.Run(where)
}

type restError struct {
	Error string `json:"error"`
}

type restValidationError struct {
	ValidationErrors []string `json:"validation_errors"`
}

// e writes err to the caller, with the given HTTP status.
func e(c *gin.Context, status int, err error) {
	var jsonErr interface{}

	jsonErr = restError{
		Error: err.Error(),
	}

	ve := validator.ValidationErrors{}
	if errors.As(err, &ve) {
		rve := restValidationError{}
		for _, v := range ve {
			rve.ValidationErrors = append(rve.ValidationErrors, v.Error())
		}

		jsonErr = rve
	}

	_ = c.Error(err)
	c.AbortWithStatusJSON(status, jsonErr)
}
