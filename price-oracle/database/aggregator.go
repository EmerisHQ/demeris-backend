package database

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/zap"

	"github.com/allinbits/demeris-backend/price-oracle/config"
	"github.com/allinbits/demeris-backend/price-oracle/types"
)

func StartAggregate(ctx context.Context, logger *zap.SugaredLogger, cfg *config.Config) {

	d, err := New(cfg.DatabaseConnectionURL)
	if err != nil {
		logger.Fatal(err)
	}
	defer d.d.Close()

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		AggregateWokers(ctx, d.d.DB, logger, cfg, PricetokenAggregator)
	}()
	go func() {
		defer wg.Done()
		AggregateWokers(ctx, d.d.DB, logger, cfg, PricefiatAggregator)
	}()

	wg.Wait()
}

func AggregateWokers(ctx context.Context, db *sqlx.DB, logger *zap.SugaredLogger, cfg *config.Config, fn func(context.Context, *sqlx.DB, *zap.SugaredLogger, *config.Config) error) {
	logger.Infow("INFO", "DB", "Aggregate WORK Start")
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if err := fn(ctx, db, logger, cfg); err != nil {
			logger.Errorw("DB", "Aggregate WORK err", err)
		}

		interval, err := time.ParseDuration(cfg.Interval)
		if err != nil {
			logger.Errorw("DB", "Aggregate WORK err", err)
			return
		}
		time.Sleep(interval)
	}
}

func PricetokenAggregator(ctx context.Context, db *sqlx.DB, logger *zap.SugaredLogger, cfg *config.Config) error {
	symbolkv := make(map[string][]float64)
	var query []string
	binanceQuery := "SELECT * FROM oracle.binance"
	//coinmarketcapQuery := "SELECT * FROM oracle.coinmarketcap"
	coinmarketgeckoQuery := "SELECT * FROM oracle.coingecko"
	query = append(query, binanceQuery)
	query = append(query, coinmarketgeckoQuery)

	whitelist := make(map[string]struct{})
	cnswhitelist, err := CnsTokenQuery(db)
	if err != nil {
		return fmt.Errorf("CnsTokenQuery: %w", err)
	}
	for _, token := range cnswhitelist {
		basetoken := token + types.USDTBasecurrency
		whitelist[basetoken] = struct{}{}
	}

	for _, q := range query {
		Prices := PriceQuery(db, logger, q)
		for _, apitokenList := range Prices {
			if _, ok := whitelist[apitokenList.Symbol]; !ok {
				continue
			}
			now := time.Now()
			if apitokenList.UpdatedAt < now.Unix()-60 {
				continue
			}
			Pricelist := symbolkv[apitokenList.Symbol]
			Pricelist = append(Pricelist, apitokenList.Price)
			symbolkv[apitokenList.Symbol] = Pricelist
		}
	}

	for token := range whitelist {
		var total float64 = 0
		for _, value := range symbolkv[token] {
			total += value
		}
		if len(symbolkv[token]) == 0 {
			return nil
		}

		median := total / float64(len(symbolkv[token]))
		tx := db.MustBegin()

		result := tx.MustExec("UPDATE oracle.tokens SET price = ($1) WHERE symbol = ($2)", median, token)
		updateresult, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("DB update: %w", err)
		}
		//If you perform an update without a token column, it does not respond as an error; it responds with zero.
		//So you have to insert a new one in the column.
		if updateresult == 0 {
			tx.MustExec("INSERT INTO oracle.tokens VALUES (($1),($2));", token, median)
		}

		err = tx.Commit()
		if err != nil {
			return fmt.Errorf("DB commit: %w", err)
		}
		logger.Infow("Insert to median Token Price", token, median)
	}
	return nil
}

func PricefiatAggregator(ctx context.Context, db *sqlx.DB, logger *zap.SugaredLogger, cfg *config.Config) error {
	symbolkv := make(map[string][]float64)
	var query []string
	fixerQuery := "SELECT * FROM oracle.fixer"

	query = append(query, fixerQuery)
	whitelist := make(map[string]struct{})
	for _, fiat := range cfg.Whitelistfiats {
		basefiat := types.USDBasecurrency + fiat
		whitelist[basefiat] = struct{}{}
	}

	for _, q := range query {
		Prices := PriceQuery(db, logger, q)
		for _, apifiatList := range Prices {
			if _, ok := whitelist[apifiatList.Symbol]; !ok {
				continue
			}
			now := time.Now()
			if apifiatList.UpdatedAt < now.Unix()-60 {
				continue
			}
			Pricelist := symbolkv[apifiatList.Symbol]
			Pricelist = append(Pricelist, apifiatList.Price)
			symbolkv[apifiatList.Symbol] = Pricelist
		}
	}
	for fiat := range whitelist {
		var total float64 = 0
		for _, value := range symbolkv[fiat] {
			total += value
		}
		if len(symbolkv[fiat]) == 0 {
			return nil
		}
		median := total / float64(len(symbolkv[fiat]))

		tx := db.MustBegin()

		result := tx.MustExec("UPDATE oracle.fiats SET price = ($1) WHERE symbol = ($2)", median, fiat)
		updateresult, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("DB update: %w", err)
		}
		if updateresult == 0 {
			tx.MustExec("INSERT INTO oracle.fiats VALUES (($1),($2));", fiat, median)
		}
		err = tx.Commit()
		if err != nil {
			return fmt.Errorf("DB commit: %w", err)
		}
		logger.Infow("Insert to median Fiat Price", fiat, median)
	}
	return nil
}

func PriceQuery(db *sqlx.DB, logger *zap.SugaredLogger, Query string) []types.Prices {
	var symbols []types.Prices
	var symbol types.Prices
	rows, err := db.Queryx(Query)
	if err != nil {
		logger.Fatalw("Fatal", "DB", err.Error(), "Duration", time.Second)
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.StructScan(&symbol)
		if err != nil {
			logger.Fatalw("Fatal", "DB", err.Error(), "Duration", time.Second)
		}
		symbols = append(symbols, symbol)
	}
	return symbols
}
