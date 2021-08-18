package rest

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/allinbits/demeris-backend/price-oracle/types"
	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v4/stdlib"
)

const getselectTokensPricesRoute = "/tokens"

func selectTokensPrices(r *router, selectToken types.SelectToken) ([]types.TokenPriceResponse, error) {
	var Tokens []types.TokenPriceResponse
	var Token types.TokenPriceResponse
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
	defer rows.Close()
	for rows.Next() {
		var symbol string
		var price float64
		var supply float64
		err := rows.Scan(&symbol, &price)
		if err != nil {
			r.s.l.Error("Error", "DB", err.Error(), "Duration", time.Second)
			return nil, err
		}
		//rowCmcSupply, err := r.s.d.Query("SELECT * FROM oracle.coinmarketcapsupply WHERE symbol=$1", symbol)
		rowCmcSupply, err := r.s.d.Query("SELECT * FROM oracle.coingeckosupply WHERE symbol=$1", symbol)
		if err != nil {
			r.s.l.Error("Error", "DB", err.Error(), "Duration", time.Second)
			return nil, err
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

	return Tokens, nil
}

func (r *router) TokensPrices(ctx *gin.Context) {
	var selectToken types.SelectToken
	var symbols []types.TokenPriceResponse

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
		tokens := token + types.USDTBasecurrency
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
	selectTokenkey, err := json.Marshal(selectToken.Tokens)
	if r.s.ri.Exists(string(selectTokenkey)) {
		bz, err := r.s.ri.Client.Get(context.Background(), string(selectTokenkey)).Bytes()
		if err != nil {
			r.s.l.Error("Error", "Redis-Get", err.Error(), "Duration", time.Second)
			return
		}
		err = json.Unmarshal(bz, &symbols)
		if err != nil {
			r.s.l.Error("Error", "Redis-Unmarshal", err.Error(), "Duration", time.Second)
			return
		}
		ctx.JSON(http.StatusOK, gin.H{
			"status":  http.StatusOK,
			"data":    &symbols,
			"message": nil,
		})

		return
	}
	symbols, err = selectTokensPrices(r, selectToken)
	if err != nil {
		e(ctx, http.StatusInternalServerError, err)
		return
	}
	bz, err := json.Marshal(symbols)
	err = r.s.ri.SetWithExpiryTime(string(selectTokenkey), string(bz), 10*time.Second)
	if err != nil {
		r.s.l.Error("Error", "Redis-Set", err.Error(), "Duration", time.Second)
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
