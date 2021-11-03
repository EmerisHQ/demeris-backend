package rest

import (
	"errors"
	"net/http"

	"github.com/allinbits/demeris-backend/utils/validation"

	"github.com/allinbits/demeris-backend/utils/k8s"

	"github.com/gin-gonic/gin"
)

const deleteChainRoute = "/delete"

type deleteChainRequest struct {
	Chain string `json:"chain" binding:"required"`
}

func (r *router) deleteChainHandler(ctx *gin.Context) {
	chain := deleteChainRequest{}

	if err := ctx.ShouldBindJSON(&chain); err != nil {
		e(ctx, http.StatusBadRequest, validation.MissingFieldsErr(err, false))
		r.s.l.Error("cannot bind input data to Chain struct", err)
		return
	}

	k := k8s.Querier{Client: *r.s.k, Namespace: r.s.defaultK8SNamespace}

	if err := k.DeleteNode(chain.Chain); err != nil {
		// there isn't always a k8s nodeset for a given chain
		if !errors.Is(err, k8s.ErrNotFound) {
			e(ctx, http.StatusInternalServerError, err)
			r.s.l.Error("cannot delete chain", err)
			return
		}
	}

	if err := r.s.d.DeleteChain(chain.Chain); err != nil {
		e(ctx, http.StatusInternalServerError, err)
		r.s.l.Error("cannot delete chain", err)
		return
	}

	return
}
func (r *router) deleteChain() (string, gin.HandlerFunc) {
	return deleteChainRoute, r.deleteChainHandler
}