package rest

import (
	"net/http"
	"strconv"
	"time"

	"github.com/allinbits/demeris-backend/price-oracle/types"
	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v4/stdlib"
)

const getselectFiatsPricesRoute = "/fiats"

func selectFiatsPrices(r *router, selectFiat types.SelectFiat) ([]types.ResponsePrices, error) {
	var symbols []types.ResponsePrices
	var symbol types.ResponsePrices
	var symbolList []interface{}

	symbolNum := len(selectFiat.Fiats)

	query := "SELECT * FROM oracle.fiats WHERE symbol=$1"

	for i := 2; i <= symbolNum; i++ {
		query += " OR" + " symbol=$" + strconv.Itoa(i)
	}

	for _, usersymbol := range selectFiat.Fiats {
		symbolList = append(symbolList, usersymbol)
	}

	rows, err := r.s.d.Query(query, symbolList...)
	if err != nil {
		r.s.l.Error("Error", "DB", err.Error(), "Duration", time.Second)
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

func (r *router) FiatsPrices(ctx *gin.Context) {
	var selectFiat types.SelectFiat
	var symbols []types.ResponsePrices

	err := ctx.BindJSON(&selectFiat)
	if err != nil {
		r.s.l.Error("Error", "FiatsPrices", err.Error(), "Duration", time.Second)
	}
	if len(selectFiat.Fiats) >= 10 {
		ctx.JSON(http.StatusForbidden, gin.H{
			"status":  http.StatusForbidden,
			"data":    "Not allow More than 10 asset",
			"message": nil,
		})
	}
	if len(selectFiat.Fiats) == 0 {
		ctx.JSON(http.StatusForbidden, gin.H{
			"status":  http.StatusForbidden,
			"data":    "",
			"message": "Not allow 0  asset",
		})
		return
	}

	if selectFiat.Fiats == nil {
		ctx.JSON(http.StatusForbidden, gin.H{
			"status":  http.StatusForbidden,
			"data":    "",
			"message": "Not allow nil asset",
		})
		return
	}
	var basefiats []string
	for _, fiat := range r.s.c.Whitelistfiats {
		fiats := types.FiatBasecurrency + fiat
		basefiats = append(basefiats, fiats)
	}
	if Diffpair(selectFiat.Fiats, basefiats) == false {
		ctx.JSON(http.StatusForbidden, gin.H{
			"status":  http.StatusForbidden,
			"data":    "",
			"message": "Not whitelisting asset",
		})
		return
	}
	symbols, err = selectFiatsPrices(r, selectFiat)
	if err != nil {
		r.s.l.Error("Error", "SelectFiatQuery", err.Error(), "Duration", time.Second)
	}
	ctx.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"data":    &symbols,
		"message": nil,
	})
}

func (r *router) getselectFiatsPrices() (string, gin.HandlerFunc) {
	return getselectFiatsPricesRoute, r.FiatsPrices
}
