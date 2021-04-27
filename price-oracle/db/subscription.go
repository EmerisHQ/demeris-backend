package db

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"time"

	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"github.com/allinbits/demeris-backend/price-oracle/config"
	"github.com/allinbits/demeris-backend/price-oracle/server"
)

func StartSubscription(wg *sync.WaitGroup, ctx context.Context, logger *zap.Logger) {

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
	go SubscriptionWorker(wg, ctx, db, logger, SubscriptionBinance)
	wg.Add(1)
	go SubscriptionWorker(wg, ctx, db, logger, SubscriptionCoinmarketcap)
}

func SubscriptionWorker(wg *sync.WaitGroup, ctx context.Context, db *sqlx.DB, logger *zap.Logger, fn func(context.Context, *sqlx.DB, *zap.Logger) error) {
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

func SubscriptionBinance(ctx context.Context, db *sqlx.DB, logger *zap.Logger) error {
	client := http.Client{
		Timeout: 2 * time.Second,
	}
	req, err := http.NewRequest("GET", config.Config.APIs.Atom.Usd.Binance, nil)
	if err != nil {
		return fmt.Errorf("fetch binance: %w", err)
	}
	q := url.Values{}
	q.Add("symbol", "ATOMUSDT")
	req.Header.Set("Accepts", "application/json")
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)

	if err != nil {
		return fmt.Errorf("fetch binance: %w", err)
	}

	defer func() {
		resp.Body.Close()
	}()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}
	bp := server.Binance{}
	err = json.Unmarshal(body, &bp)
	if err != nil {
		return fmt.Errorf("unmarshal body: %w", err)
	}

	logger.Info("BinanceSubscription",
		zap.String("Insert to unmarshal json Price", bp.Price),
	)

	tx := db.MustBegin()
	//time.now add insert
	tx.MustExec("UPDATE oracle.binance SET price = ($2) WHERE symbol = ($1)", bp.Symbol, bp.Price)
	//https://www.cockroachlabs.com/docs/v20.2/alter-primary-key.html#alter-a-single-column-primary-key
	//https://www.cockroachlabs.com/docs/v20.2/update-data.html?filters=go#update-example
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("DB commit: %w", err)
	}
	return nil

}

func SubscriptionCoinmarketcap(ctx context.Context, db *sqlx.DB, logger *zap.Logger) error {
	client := &http.Client{
		Timeout: 2 * time.Second,
	}
	req, err := http.NewRequest("GET", config.Config.APIs.Atom.Usd.Coinmarketcap, nil)
	if err != nil {
		return fmt.Errorf("fetch coinmarketcap: %w", err)
	}
	q := url.Values{}
	q.Add("symbol", "ATOM")
	q.Add("convert", "USDT")
	req.Header.Set("Accepts", "application/json")
	req.Header.Add("X-CMC_PRO_API_KEY", config.Config.CoinmarketcapapiKey)
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("fetch coinmaketca: %w", err)
	}
	defer func() {
		resp.Body.Close()
	}()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read body: %w", err)
	}
	bp := server.Coinmarketcap{}
	err = json.Unmarshal(body, &bp)
	if err != nil {
		return fmt.Errorf("unmarshal body: %w", err)
	}
	s := fmt.Sprintf("%f", bp.Data.Atom.Quote.Usdt.Price)

	logger.Info("CoinmarketcapSubscription",
		zap.String("Insert to unmarshal json Price", s),
	)
	tx := db.MustBegin()
	tx.MustExec("UPDATE oracle.coinmarketcap SET price = ($1) WHERE symbol = 'ATOMUSDT'", s)
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("DB commit: %w", err)
	}

	return nil
}

func SubscriptionCurrencylayer() {
}

func SubscriptionBand() {
}
