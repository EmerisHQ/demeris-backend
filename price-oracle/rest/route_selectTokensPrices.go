package rest

import (
	"net/http"
	"strconv"
	"time"

	"github.com/allinbits/demeris-backend/price-oracle/types"
	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v4/stdlib"
)

const getselectTokensPricesRoute = "/tokens"

func selectTokensPrices(r *router, selectToken types.SelectToken) ([]types.ResponsePrices, error) {
	var symbols []types.ResponsePrices
	var symbol types.ResponsePrices
	var symbolList []interface{}

	symbolNum := len(selectToken.Tokens)

	query := "SELECT * FROM oracle.tokens WHERE symbol=$1"

	for i := 2; i <= symbolNum; i++ {
		query += " OR" + " symbol=$" + strconv.Itoa(i)
	}

	for _, usersymbol := range selectToken.Tokens {
		symbolList = append(symbolList, usersymbol)
	}

	rows, err := r.s.d.Query(query, symbolList...)
	if err != nil {
		r.s.l.Error("Error", "DB", err.Error(), "Duration", time.Second)
		return nil, err
	}
	for rows.Next() {
		err := rows.StructScan(&symbol)
		if err != nil {
			r.s.l.Error("Error", "DB", err.Error(), "Duration", time.Second)
			return nil, err
		}
		symbols = append(symbols, symbol)
	}

	return symbols, nil
}

func (r *router) TokensPrices(ctx *gin.Context) {
	var selectToken types.SelectToken
	var symbols []types.ResponsePrices

	err := ctx.BindJSON(&selectToken)
	if err != nil {
		r.s.l.Error("Error", "TokensPrices", err.Error(), "Duration", time.Second)
	}
	if len(selectToken.Tokens) >= 10 {
		ctx.JSON(http.StatusForbidden, gin.H{
			"status":  http.StatusForbidden,
			"data":    "",
			"message": "Not allow More than 10 asset",
		})
		return
	}

	if len(selectToken.Tokens) == 0 {
		ctx.JSON(http.StatusForbidden, gin.H{
			"status":  http.StatusForbidden,
			"data":    "",
			"message": "Not allow 0  asset",
		})
		return
	}

	if selectToken.Tokens == nil {
		ctx.JSON(http.StatusForbidden, gin.H{
			"status":  http.StatusForbidden,
			"data":    "",
			"message": "Not allow nil asset",
		})
		return
	}
	Whitelists, err := r.s.d.CnstokenQueryHandler()
	if err != nil {
		r.s.l.Error("Error", "DB", err.Error(), "Duration", time.Second)
		return
	}
	var basetokens []string
	for _, token := range Whitelists {
		tokens := token + types.TokenBasecurrency
		basetokens = append(basetokens, tokens)
	}
	if Diffpair(selectToken.Tokens, basetokens) == false {
		ctx.JSON(http.StatusForbidden, gin.H{
			"status":  http.StatusForbidden,
			"data":    "",
			"message": "Not whitelisting asset",
		})
		return
	}
	symbols, err = selectTokensPrices(r, selectToken)
	if err != nil {
		e(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"data":    &symbols,
		"message": nil,
	})
}

func (r *router) getselectTokensPrices() (string, gin.HandlerFunc) {
	return getselectTokensPricesRoute, r.TokensPrices
}
