package rest

import (
	"net/http"

	navigator_cns "github.com/allinbits/navigator-cns"
	"github.com/gin-gonic/gin"
)

const addChainRoute = "/add"

func (r *router) addChainHandler(ctx *gin.Context) {
	newChain := navigator_cns.Chain{}

	if err := ctx.ShouldBindJSON(&newChain); err != nil {
		e(ctx, http.StatusBadRequest, err)
		return
	}

	if err := r.s.d.AddChain(newChain); err != nil {
		e(ctx, http.StatusInternalServerError, err)
		return
	}

	return
}
func (r *router) addChain() (string, gin.HandlerFunc) {
	return addChainRoute, r.addChainHandler
}
