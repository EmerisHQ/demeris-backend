package router

import (
	"strings"

	"github.com/gin-gonic/gin"
	ginprometheus "github.com/zsais/go-gin-prometheus"
)

func (r *Router) metrics() {
	p := ginprometheus.NewPrometheus("demeris_api")

	p.ReqCntURLLabelMappingFn = func(c *gin.Context) string {
		url := c.Request.URL.Path
		for _, p := range c.Params {
			url = strings.Replace(url, p.Value, ":"+p.Key, 1)
		}
		return url
	}

	p.Use(r.g)
}
