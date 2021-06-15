package rest

import (
	"net/http"
	"time"

	"github.com/allinbits/demeris-backend/price-oracle/types"
	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v4/stdlib"
)

const getAllPriceRoute = "/prices"

func allPrices(r *router) ([]types.ResponsePrices, error) {
	var symbols []types.ResponsePrices
	var symbolToken types.ResponsePrices
	var symbolFiat types.ResponsePrices

	rowsToken, err := r.s.d.Query("SELECT * FROM oracle.tokens")
	if err != nil {
		r.s.l.Error("Error", "DB", err.Error(), "Duration", time.Second)
		return nil, err
	}
	Whitelists, err := r.s.d.CnstokenQueryHandler()
	if err != nil {
		r.s.l.Error("Error", "DB", err.Error(), "Duration", time.Second)
		return nil, err
	}
	for rowsToken.Next() {
		err := rowsToken.StructScan(&symbolToken)
		if err != nil {
			r.s.l.Fatalw("Error", "DB", err.Error(), "Duration", time.Second)
			return nil, err
		}
		for _, token := range Whitelists {
			token = token + types.TokenBasecurrency
			if symbolToken.Symbol == token {
				symbols = append(symbols, symbolToken)
			}
		}
	}

	rowsFiat, err := r.s.d.Query("SELECT * FROM oracle.fiats")
	if err != nil {
		r.s.l.Fatalw("Error", "DB", err.Error(), "Duration", time.Second)
		return nil, err
	}
	for rowsFiat.Next() {
		err := rowsFiat.StructScan(&symbolFiat)
		if err != nil {
			r.s.l.Errorw("Error", "DB", err.Error(), "Duration", time.Second)
			return nil, err
		}
		for _, fiat := range r.s.c.Whitelistfiats {
			fiat = types.FiatBasecurrency + fiat
			if symbolFiat.Symbol == fiat {
				symbols = append(symbols, symbolFiat)
			}
		}
	}
	return symbols, nil
}

func (r *router) allPricesHandler(ctx *gin.Context) {
	Allsymbols, err := allPrices(r)
	if err != nil {
		e(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"data":    &Allsymbols,
		"message": nil,
	})
}

func (r *router) getallPrices() (string, gin.HandlerFunc) {
	return getAllPriceRoute, r.allPricesHandler
}
