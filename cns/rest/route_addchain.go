package rest

import (
	"net/http"

	"github.com/allinbits/demeris-backend/cns/k8s"

	"github.com/allinbits/demeris-backend/models"
	"github.com/allinbits/demeris-backend/utils/k8s/operator"
	"github.com/gin-gonic/gin"
)

const addChainRoute = "/add"

type addChainRequest struct {
	models.Chain

	NodeConfig operator.NodeConfiguration `json:"node_config" binding:"required,dive"`
}

func (r *router) addChainHandler(ctx *gin.Context) {
	newChain := addChainRequest{}

	if err := ctx.ShouldBindJSON(&newChain); err != nil {
		e(ctx, http.StatusBadRequest, err)
		r.s.l.Error("cannot bind input data to Chain struct", err)
		return
	}

	k := k8s.Querier{Client: *r.s.k}

	node, err := operator.NewNode(newChain.NodeConfig)
	if err != nil {
		e(ctx, http.StatusBadRequest, err)
		r.s.l.Error("cannot add chain", err)
		return
	}

	if err := k.AddNode(*node); err != nil {
		e(ctx, http.StatusInternalServerError, err)
		r.s.l.Error("cannot add chain", err)
		return
	}

	if err := r.s.d.AddChain(newChain.Chain); err != nil {
		e(ctx, http.StatusInternalServerError, err)
		r.s.l.Error("cannot add chain", err)
		return
	}

	return
}
func (r *router) addChain() (string, gin.HandlerFunc) {
	return addChainRoute, r.addChainHandler
}
