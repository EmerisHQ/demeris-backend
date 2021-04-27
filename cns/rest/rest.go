package rest

import (
	"errors"
	"net/http"

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
}

type router struct {
	s *Server
}

func NewServer(l *zap.SugaredLogger, d *database.Instance, debug bool) *Server {
	if !debug {
		gin.SetMode(gin.ReleaseMode)
	}

	g := gin.New()

	s := &Server{
		l: l,
		d: d,
		g: g,
	}

	r := &router{s: s}

	g.Use(logging.LogRequest(l.Desugar()))
	g.Use(ginzap.RecoveryWithZap(l.Desugar(), true))

	g.GET(r.getChains())
	g.POST(r.addChain())

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

	c.Error(err)
	c.AbortWithStatusJSON(status, jsonErr)
}
