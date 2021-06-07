package rest

import (
	"fmt"
	"net/http"

	"github.com/allinbits/demeris-backend/models"

	"github.com/gin-gonic/gin"
)

const getChainRoute = "/chain/:chain"

type getChainResp struct {
	Chain models.Chain `json:"chain"`
}

func (r *router) getChainHandler(ctx *gin.Context) {

	chain, ok := ctx.Params.Get("chain")

	if !ok {
		e(ctx, http.StatusBadRequest, fmt.Errorf("chain not supplied"))
	}

	data, err := r.s.d.Chain(chain)

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
