package rest

import (
	"context"
	"net/http"

	cnsdb2 "github.com/allinbits/demeris-backend/cns/cnsdb"
	"github.com/gin-gonic/gin"
)

const getChainsRoute = "/chains"

type getChainsResp struct {
	Chains []cnsdb2.Chain `json:"chains"`
}

func (r *router) getChainsHandler(ctx *gin.Context) {
	data, err := r.s.d.Chains(context.Background())

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
