package rest

import (
	"net/http"

	demeris_cns "github.com/allinbits/demeris-backend/cns"
	"github.com/gin-gonic/gin"
)

const addChainRoute = "/add"

func (r *router) addChainHandler(ctx *gin.Context) {
	newChain := demeris_cns.Chain{}

	if err := ctx.ShouldBindJSON(&newChain); err != nil {
		e(ctx, http.StatusBadRequest, err)
		r.s.l.Error("cannot bind input data to Chain struct", err)
		return
	}

	if err := r.s.d.AddChain(newChain); err != nil {
		e(ctx, http.StatusInternalServerError, err)
		r.s.l.Error("cannot add chain", err)
		return
	}

	return
}
func (r *router) addChain() (string, gin.HandlerFunc) {
	return addChainRoute, r.addChainHandler
}
