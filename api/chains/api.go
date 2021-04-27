package chains

import (
	"fmt"
	"net/http"

	"github.com/allinbits/demeris-backend/api/router/deps"
	"github.com/gin-gonic/gin"
)

func Register(router *gin.Engine) {
	router.GET("/chains", GetChains)
}

// GetVerifiedDenoms returns the list of all the chains supported by demeris.
func GetChains(c *gin.Context) {
	var res chainsResponse

	d, err := deps.GetDeps(c)
	if err != nil {
		c.Error(deps.NewError(
			"chains",
			fmt.Errorf("internal error"),
			http.StatusInternalServerError,
		))

		panic("cannot retrieve context deps")
		return
	}

	chains, err := d.Database.Chains()

	if err != nil {
		e := deps.NewError(
			"chains",
			fmt.Errorf("cannot retrieve chains"),
			http.StatusBadRequest,
		)

		c.Error(e)

		d.Logger.Errorw(
			"cannot retrieve chains",
			"id",
			e.ID,
			"error",
			err,
		)

		return
	}

	for _, cc := range chains {
		res.SupportedChains = append(res.SupportedChains, cc.ChainName)
	}

	c.JSON(http.StatusOK, res)
}
