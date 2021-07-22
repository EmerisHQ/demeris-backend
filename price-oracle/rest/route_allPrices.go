package rest

import (
	"net/http"
	"time"

	"github.com/allinbits/demeris-backend/price-oracle/types"
	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v4/stdlib"
)

const getAllPriceRoute = "/prices"

func allPrices(r *router) ([]types.TokenPriceResponse, []types.FiatPriceResponse, error) {
	var Fiats []types.FiatPriceResponse
	var Fiat types.FiatPriceResponse
	var Tokens []types.TokenPriceResponse
	var Token types.TokenPriceResponse

	rowsToken, err := r.s.d.Query("SELECT * FROM oracle.tokens")
	if err != nil {
		r.s.l.Error("Error", "DB", err.Error(), "Duration", time.Second)
		return nil, nil, err
	}
	defer rowsToken.Close()
	Whitelists, err := r.s.d.CnstokenQueryHandler()
	if err != nil {
		r.s.l.Error("Error", "DB", err.Error(), "Duration", time.Second)
		return nil, nil, err
	}
	for rowsToken.Next() {
		var symbol string
		var price float64
		var supply float64
		err := rowsToken.Scan(&symbol, &price)
		if err != nil {
			r.s.l.Fatalw("Error", "DB", err.Error(), "Duration", time.Second)
			return nil, nil, err
		}
		for _, Whitelisttoken := range Whitelists {
			Whitelisttoken = Whitelisttoken + types.TokenBasecurrency
			if symbol == Whitelisttoken {
				rowCmcSupply, err := r.s.d.Query("SELECT * FROM oracle.coinmarketcapsupply WHERE symbol=$1", Whitelisttoken)
				if err != nil {
					r.s.l.Error("Error", "DB", err.Error(), "Duration", time.Second)
					return nil, nil, err
				}
				defer rowCmcSupply.Close()
				for rowCmcSupply.Next() {
					if err := rowCmcSupply.Scan(&symbol, &supply); err != nil {
						r.s.l.Error("Error", "DB", err.Error(), "Duration", time.Second)
					}
				}
				Token.Symbol = symbol
				Token.Price = price
				Token.Supply = supply
				Tokens = append(Tokens, Token)
			}
		}
	}

	rowsFiat, err := r.s.d.Query("SELECT * FROM oracle.fiats")
	if err != nil {
		r.s.l.Fatalw("Error", "DB", err.Error(), "Duration", time.Second)
		return nil, nil, err
	}
	defer rowsFiat.Close()
	for rowsFiat.Next() {
		var symbol string
		var price float64
		err := rowsFiat.Scan(&symbol, &price)
		if err != nil {
			r.s.l.Errorw("Error", "DB", err.Error(), "Duration", time.Second)
			return nil, nil, err
		}
		for _, fiat := range r.s.c.Whitelistfiats {
			fiat = types.FiatBasecurrency + fiat

			Fiat.Symbol = symbol
			Fiat.Price = price
			if symbol == fiat {
				Fiats = append(Fiats, Fiat)
			}
		}
	}

	return Tokens, Fiats, nil
}

func (r *router) allPricesHandler(ctx *gin.Context) {
	var AllPriceResponse types.AllPriceResponse
	Tokens, Fiats, err := allPrices(r)
	AllPriceResponse.Tokens = Tokens
	AllPriceResponse.Fiats = Fiats
	if err != nil {
		e(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"data":    &AllPriceResponse,
		"message": nil,
	})
}

func (r *router) getallPrices() (string, gin.HandlerFunc) {
	return getAllPriceRoute, r.allPricesHandler
}
