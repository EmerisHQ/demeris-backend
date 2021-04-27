package rest

import (
	"net/http"

	demeris_cns "github.com/allinbits/demeris-backend/cns"
	"github.com/gin-gonic/gin"
)

const getChainsRoute = "/chains"

type getChainsResp struct {
	Chains []demeris_cns.Chain `json:"chains"`
}

func (r *router) getChainsHandler(ctx *gin.Context) {
	data, err := r.s.d.Chains()

	if err != nil {
		e(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, getChainsResp{
		Chains: data,
	})
}
func (r *router) getChains() (string, gin.HandlerFunc) {
	return getChainsRoute, r.getChainsHandler
}
