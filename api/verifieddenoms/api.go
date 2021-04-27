package verifieddenoms

import (
	"fmt"
	"net/http"

	"github.com/allinbits/demeris-backend/api/router/deps"
	"github.com/gin-gonic/gin"
)

func Register(router *gin.Engine) {
	router.GET("/verified_denoms", GetVerifiedDenoms)
}

// GetVerifiedDenoms returns the fee token for a given chain, looked up by the chain name attribute.
func GetVerifiedDenoms(c *gin.Context) {
	var res verifiedDenomsResponse

	d, err := deps.GetDeps(c)
	if err != nil {
		c.Error(deps.NewError(
			"verified_denoms",
			fmt.Errorf("internal error"),
			http.StatusInternalServerError,
		))

		panic("cannot retrieve context deps")
		return
	}

	chains, err := d.Database.Chains()

	if err != nil {
		e := deps.NewError(
			"verified_denoms",
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
		for _, vd := range cc.VerifiedNativeDenoms() {
			res.VerifiedDenoms = append(res.VerifiedDenoms, VerifiedDenom{
				Name:     vd.Name,
				Verified: vd.Verified,
			})
		}
	}

	c.JSON(http.StatusOK, res)
}
