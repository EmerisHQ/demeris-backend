package rest

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"

	"github.com/allinbits/navigator-cns/database"
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

	g.Use(s.logReq())
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

// logReq is a middleware which logs each requests as they come.
func (s *Server) logReq() gin.HandlerFunc {
	if s.gl == nil {
		s.gl = s.l.Desugar()
	}

	return func(c *gin.Context) {
		start := time.Now()
		// some evil middlewares modify this values
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		s.gl.Info(path,
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.String("user-agent", c.Request.UserAgent()),
			zap.String("time", start.Format(time.RFC3339)),
		)

		c.Next()

		if len(c.Errors) > 0 {
			// Append error field if this is an erroneous request.
			for _, e := range c.Errors.Errors() {
				s.gl.Error(e)
			}
		}
	}
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
