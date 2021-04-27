package server

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"github.com/allinbits/demeris-backend/price-oracle/config"
)

var db *sqlx.DB

func ConnectDB(wg *sync.WaitGroup) {
	dbLocal, err := sqlx.Connect("pgx", config.Config.DB)
	db = dbLocal

	if err != nil {
		wg.Done()
		Logger.Error("Error",
			zap.String("DB", err.Error()),
			zap.Duration("Duration", time.Second),
		)
	}

	Logger.Info("INFO",
		zap.String("DB", "Connect: "+config.Config.DB),
		zap.Duration("Duration", time.Second),
	)
}

func AllPriceQuery() []Prices {
	var symbols []Prices
	var symbolToken Prices
	var symbolFiat Prices

	rowsToken, err := db.Queryx("SELECT * FROM oracle.tokens")
	if err != nil {
		Logger.Error("Error",
			zap.String("DB", err.Error()),
			zap.Duration("Duration", time.Second),
		)
	}
	for rowsToken.Next() {
		err := rowsToken.StructScan(&symbolToken)
		if err != nil {
			Logger.Fatal("Error",
				zap.String("DB", err.Error()),
				zap.Duration("Duration", time.Second),
			)
		}
		symbols = append(symbols, symbolToken)
	}

	rowsFiat, err := db.Queryx("SELECT * FROM oracle.fiats")
	if err != nil {
		Logger.Fatal("Error",
			zap.String("DB", err.Error()),
			zap.Duration("Duration", time.Second),
		)
	}
	for rowsFiat.Next() {
		err := rowsFiat.StructScan(&symbolFiat)
		if err != nil {
			Logger.Error("Error",
				zap.String("DB", err.Error()),
				zap.Duration("Duration", time.Second),
			)
		}
		symbols = append(symbols, symbolFiat)
	}
	return symbols
}

func SelectPriceQuery(selectToken SelectToken) []Prices {
	var symbols []Prices
	var symbol Prices
	var symbolList []interface{}

	symbolNum := len(selectToken.Tokens)

	query := "SELECT * FROM oracle.tokens WHERE symbol=$1"

	for i := 2; i <= symbolNum; i++ {
		query += " OR" + " symbol=$" + strconv.Itoa(i)
	}

	for _, usersymbol := range selectToken.Tokens {
		symbolList = append(symbolList, usersymbol)
	}

	rows, err := db.Queryx(query, symbolList...)
	if err != nil {
		Logger.Error("Error",
			zap.String("DB", err.Error()),
			zap.Duration("Duration", time.Second),
		)
	}
	for rows.Next() {
		err := rows.StructScan(&symbol)
		if err != nil {
			Logger.Error("Error",
				zap.String("DB", err.Error()),
				zap.Duration("Duration", time.Second),
			)
		}
		symbols = append(symbols, symbol)
	}

	return symbols
}
func SelectFiatQuery(selectFiat SelectFiat) []Prices {
	var symbols []Prices
	var symbol Prices
	var symbolList []interface{}

	symbolNum := len(selectFiat.Fiat)

	query := "SELECT * FROM oracle.fiats WHERE symbol=$1"

	for i := 2; i <= symbolNum; i++ {
		query += " OR" + " symbol=$" + strconv.Itoa(i)
	}

	for _, usersymbol := range selectFiat.Fiat {
		symbolList = append(symbolList, usersymbol)
	}

	rows, err := db.Queryx(query, symbolList...)
	if err != nil {
		Logger.Error("Error",
			zap.String("DB", err.Error()),
			zap.Duration("Duration", time.Second),
		)
	}
	for rows.Next() {
		err := rows.StructScan(&symbol)
		if err != nil {
			Logger.Error("Error",
				zap.String("DB", err.Error()),
				zap.Duration("Duration", time.Second),
			)
		}
		symbols = append(symbols, symbol)
	}

	return symbols
}

func diffpair(a []string, b []string) bool {
	// Turn b into a map
	var m map[string]bool
	m = make(map[string]bool, len(b))
	for _, s := range b {
		m[s] = false
	}
	// Append values from the longest slice that don't exist in the map
	var diff []string
	for _, s := range a {
		if _, ok := m[s]; !ok {
			diff = append(diff, s)
			continue
		}
		m[s] = true
	}

	if diff == nil {
		return true
	}
	return false
}
func AllTokenPrices(c *gin.Context) {
	Allsymbols := AllPriceQuery()
	Logger.Info("INFO",
		zap.String("AllTokenPrices", "String"),
		zap.Duration("Duration", time.Second),
	)
	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"data":    &Allsymbols,
		"message": nil,
	})
}

func TokensPrices(c *gin.Context) {
	var selectToken SelectToken
	var symbols []Prices

	err := c.BindJSON(&selectToken)
	if err != nil {
		Logger.Error("Error",
			zap.String("TokensPrices", err.Error()),
			zap.Duration("Duration", time.Second),
		)
	}
	if len(selectToken.Tokens) >= 10 {
		c.JSON(http.StatusForbidden, gin.H{
			"status":  http.StatusForbidden,
			"data":    "",
			"message": "Not allow More than 10 asset",
		})
		return
	}

	if len(selectToken.Tokens) == 0 {
		c.JSON(http.StatusForbidden, gin.H{
			"status":  http.StatusForbidden,
			"data":    "",
			"message": "Not allow 0  asset",
		})
		return
	}

	if selectToken.Tokens == nil {
		c.JSON(http.StatusForbidden, gin.H{
			"status":  http.StatusForbidden,
			"data":    "",
			"message": "Not allow nil asset",
		})
		return
	}
	if diffpair(selectToken.Tokens, config.Config.Whitelisttokens) == false {
		c.JSON(http.StatusForbidden, gin.H{
			"status":  http.StatusForbidden,
			"data":    "",
			"message": "Not whitelisting asset",
		})
		return
	}
	symbols = SelectPriceQuery(selectToken)

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"data":    &symbols,
		"message": nil,
	})
}

func FiatsPrices(c *gin.Context) {
	var selectFiat SelectFiat
	var symbols []Prices

	err := c.BindJSON(&selectFiat)
	if err != nil {
		Logger.Error("Error",
			zap.String("FiatsPrices", err.Error()),
			zap.Duration("Duration", time.Second),
		)
	}
	if len(selectFiat.Fiat) >= 10 {
		c.JSON(http.StatusForbidden, gin.H{
			"status":  http.StatusForbidden,
			"data":    "Not allow More than 10 asset",
			"message": nil,
		})
	}
	if len(selectFiat.Fiat) == 0 {
		c.JSON(http.StatusForbidden, gin.H{
			"status":  http.StatusForbidden,
			"data":    "",
			"message": "Not allow 0  asset",
		})
		return
	}

	if selectFiat.Fiat == nil {
		c.JSON(http.StatusForbidden, gin.H{
			"status":  http.StatusForbidden,
			"data":    "",
			"message": "Not allow nil asset",
		})
		return
	}
	if diffpair(selectFiat.Fiat, config.Config.Whitelistfiats) == false {
		c.JSON(http.StatusForbidden, gin.H{
			"status":  http.StatusForbidden,
			"data":    "",
			"message": "Not whitelisting asset",
		})
		return
	}
	symbols = SelectFiatQuery(selectFiat)

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"data":    &symbols,
		"message": nil,
	})
}
