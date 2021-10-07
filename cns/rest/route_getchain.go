package rest

import (
	"context"
	"fmt"
	"net/http"

	"github.com/allinbits/demeris-backend/cns/cnsdb"
	"github.com/gin-gonic/gin"
)

const getChainRoute = "/chain/:chain"

type getChainResp struct {
	Chain cnsdb.Chain `json:"chain"`
}

func (r *router) getChainHandler(ctx *gin.Context) {

	chain, ok := ctx.Params.Get("chain")

	if !ok {
		e(ctx, http.StatusBadRequest, fmt.Errorf("chain not supplied"))
	}

	data, err := r.s.d.Chain(context.Background(), chain)

	if err != nil {
		e(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, getChainResp{
		Chain: data,
	})
}
func (r *router) getChain() (string, gin.HandlerFunc) {
	return getChainRoute, r.getChainHandler
}
