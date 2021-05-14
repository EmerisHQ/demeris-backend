package deps

import (
	"fmt"

	kube "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/allinbits/demeris-backend/api/database"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Deps represents a set of objects useful during the lifecycle of REST endpoints.
type Deps struct {
	Logger   *zap.SugaredLogger
	Database *database.Database
	K8S      *kube.Client
	CNSURL   string
}

func GetDeps(c *gin.Context) *Deps {
	d, ok := c.Get("deps")
	if !ok {
		panic("deps not set in context")
	}

	deps, ok := d.(*Deps)
	if !ok {
		panic(fmt.Sprintf("deps not of the expected type, found %T", deps))
	}

	return deps
}

func (d *Deps) WriteError(c *gin.Context, err Error, logMessage string, keyAndValues ...interface{}) {
	_ = c.Error(err)

	if keyAndValues != nil {
		d.Logger.Errorw(
			logMessage,
			keyAndValues...,
		)
	}
}
