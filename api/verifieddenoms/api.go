package verifieddenoms

import (
	"fmt"
	"net/http"
	"time"

	"github.com/allinbits/demeris-backend/api/router/deps"
	"github.com/gin-contrib/cache"
	"github.com/gin-contrib/cache/persistence"
	"github.com/gin-gonic/gin"
)

func Register(router *gin.Engine, store *persistence.InMemoryStore) {
	router.GET("/verified_denoms", cache.CachePage(store, 10*time.Second, GetVerifiedDenoms))
}

// GetVerifiedDenoms returns the list of verified denoms.
// @Summary Gets verified denoms
// @Tags Denoms
// @ID verified-denoms
// @Description gets verified denoms
// @Produce json
// @Success 200 {object} verifiedDenomsResponse
// @Failure 500,403 {object} deps.Error
// @Router /verified_denoms [get]
func GetVerifiedDenoms(c *gin.Context) {
	var res verifiedDenomsResponse

	d := deps.GetDeps(c)

	chains, err := d.Database.Chains()

	if err != nil {
		e := deps.NewError(
			"verified_denoms",
			fmt.Errorf("cannot retrieve chains"),
			http.StatusBadRequest,
		)

		d.WriteError(c, e,
			"cannot retrieve chains",
			"id",
			e.ID,
			"error",
			err,
		)

		return
	}

	for _, cc := range chains {
		for _, vd := range cc.VerifiedTokens() {
			res.VerifiedDenoms = append(res.VerifiedDenoms, verifiedDenom{
				Denom:     vd,
				ChainName: cc.ChainName,
			})
		}
	}

	c.JSON(http.StatusOK, res)
}
