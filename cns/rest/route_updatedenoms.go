package rest

import (
	"net/http"

	"github.com/allinbits/demeris-backend/models"
	"github.com/gin-gonic/gin"
)

const updateDenomsRoute = "/denoms"

type updateDenomsRequest struct {
	Chain  string           `json:"chain_name"`
	Denoms models.DenomList `json:"denoms"`
}

func (r *router) updateDenomsHandler(ctx *gin.Context) {
	req := updateDenomsRequest{}

	if err := ctx.BindJSON(&req); err != nil {
		e(ctx, http.StatusBadRequest, err)
		r.s.l.Error("cannot bind json to updateDenomsRequest", err)
		return

	}

	if err := r.s.d.UpdateDenoms(req.Chain, req.Denoms); err != nil {
		e(ctx, http.StatusInternalServerError, err)
		r.s.l.Error("cannot update denoms", err)
		return
	}

	return
}
func (r *router) updateDenoms() (string, gin.HandlerFunc) {
	return updateDenomsRoute, r.updateDenomsHandler
}
