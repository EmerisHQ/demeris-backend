package deps

import (
	"fmt"

	"github.com/allinbits/demeris-backend/api/database"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Deps represents a set of objects useful during the lifecycle of REST endpoints.
type Deps struct {
	Logger   *zap.SugaredLogger
	Database *database.Database
	CNSURL   string
}

func GetDeps(c *gin.Context) (*Deps, error) {
	d, ok := c.Get("deps")
	if !ok {
		return nil, fmt.Errorf("deps not set in context")
	}

	deps, ok := d.(*Deps)
	if !ok {
		return nil, fmt.Errorf("deps not of the expected type")
	}

	return deps, nil
}
