package db

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"

	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"github.com/allinbits/navigator-price-oracle/config"
	"github.com/allinbits/navigator-price-oracle/server"
)

func StartAggregate(wg *sync.WaitGroup, ctx context.Context, logger *zap.Logger) {

	db, err := sqlx.Connect("pgx", config.Config.DB)

	if err != nil {
		logger.Fatal("Fatal",
			zap.String("DB", err.Error()),
			zap.Duration("Duration", time.Second),
		)
	}

	logger.Info("INFO",
		zap.String("DB", "Connect: "+config.Config.DB),
		zap.Duration("Duration", time.Second),
	)
	wg.Add(1)
	go AggregateWokers(wg, ctx, db, logger, PricetokenAggregator)
	wg.Add(1)
	go AggregateWokers(wg, ctx, db, logger, PricefiatAggregator)
}

func AggregateWokers(wg *sync.WaitGroup, ctx context.Context, db *sqlx.DB, logger *zap.Logger, fn func(context.Context, *sqlx.DB, *zap.Logger) error) {
	logger.Info("INFO",
		zap.String("DB", "WORK Start"),
	)
	defer db.Close()
	defer wg.Done()
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	go func() {
		<-quit
		logger.Info("INFO",
			zap.String("DB", "receive interrupt signal/Close"),
			zap.Duration("Duration", time.Second),
		)
		wg.Done()

		if err := db.Close(); err != nil {
			logger.Fatal("Fatal",
				zap.String("DB", err.Error()),
				zap.Duration("Duration", time.Second),
			)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if err := fn(ctx, db, logger); err != nil {
			logger.Error("ERROR",
				zap.Error(err),
				zap.Duration("Duration", time.Second),
			)
		}

		time.Sleep(config.Config.Interval * time.Second)
	}
}

func PricetokenAggregator(ctx context.Context, db *sqlx.DB, logger *zap.Logger) error {
	symbolkv := make(map[string][]float64)
	var query []string
	binanceQuery := "SELECT * FROM oracle.binance"
	coinmarketcapQuery := "SELECT * FROM oracle.coinmarketcap"
	query = append(query, binanceQuery)
	query = append(query, coinmarketcapQuery)

	for _, q := range query {
		Prices := PriceQuery(db, logger, q)
		var Pricelist []float64
		for _, apitokenList := range Prices {
			for _, whitelistTokenList := range config.Config.WhitelistTokens {
				//if Exclude if insert timemp is greater than 15 seconds
				if apitokenList.Symbol == whitelistTokenList {
					strtofloat, err := strconv.ParseFloat(apitokenList.Prirce, 64)
					if err != nil {
						return fmt.Errorf("ParseFloat: %w", err)
					}
					Pricelist = append(Pricelist, strtofloat)
					symbolkv[apitokenList.Symbol] = Pricelist
				}
			}
		}
	}
	for _, whitelistTokenList := range config.Config.WhitelistTokens {
		var total float64 = 0
		for _, value := range symbolkv[whitelistTokenList] {
			total += value
		}
		median := total / float64(len(symbolkv[whitelistTokenList]))
		tx := db.MustBegin()
		s := fmt.Sprintf("%f", median)
		tx.MustExec("UPDATE aggregate.tokens SET price = ($1) WHERE symbol = ($2)", s, whitelistTokenList)
		err := tx.Commit()
		if err != nil {
			return fmt.Errorf("DB commit: %w", err)
		}
	}
	return nil
}
func PricefiatAggregator(ctx context.Context, db *sqlx.DB, logger *zap.Logger) error {
	symbolkv := make(map[string][]float64)
	var query []string
	currencylayerQuery := "SELECT * FROM oracle.currencylayer"

	query = append(query, currencylayerQuery)

	for _, q := range query {
		Prices := PriceQuery(db, logger, q)
		var Pricelist []float64
		for _, apiTokenList := range Prices {
			for _, whiteListTokenList := range config.Config.WhitelistFiats {
				//if Exclude if insert timemp is greater than 15 seconds
				if apiTokenList.Symbol == whiteListTokenList {
					strToFloat, err := strconv.ParseFloat(apiTokenList.Prirce, 64)
					if err != nil {
						return fmt.Errorf("ParseFloat: %w", err)
					}
					Pricelist = append(Pricelist, strToFloat)
					symbolkv[apiTokenList.Symbol] = Pricelist
				}
			}
		}
	}
	for _, whitelistTokenList := range config.Config.WhitelistFiats {
		var total float64 = 0
		for _, value := range symbolkv[whitelistTokenList] {
			total += value
		}
		median := total / float64(len(symbolkv[whitelistTokenList]))
		tx := db.MustBegin()
		s := fmt.Sprintf("%f", median)
		tx.MustExec("UPDATE aggregate.fiats SET price = ($1) WHERE symbol = ($2)", s, whitelistTokenList)
		err := tx.Commit()
		if err != nil {
			return fmt.Errorf("DB commit: %w", err)
		}
	}
	return nil
}

func PriceQuery(db *sqlx.DB, logger *zap.Logger, Query string) []server.Prices {
	var symbols []server.Prices
	var symbol server.Prices
	rows, err := db.Queryx(Query)
	if err != nil {
		logger.Fatal("Fatal",
			zap.String("DB", err.Error()),
			zap.Duration("Duration", time.Second),
		)
	}
	for rows.Next() {
		err := rows.StructScan(&symbol)
		if err != nil {
			logger.Fatal("Fatal",
				zap.String("DB", err.Error()),
				zap.Duration("Duration", time.Second),
			)
		}
		symbols = append(symbols, symbol)
	}
	return symbols
}
