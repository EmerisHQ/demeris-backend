package database

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"github.com/allinbits/demeris-backend/price-oracle/config"
	"github.com/allinbits/demeris-backend/price-oracle/types"
)

const (
	BinanceURL       = "https://api.binance.com/api/v3/ticker/price"
	CoinmarketcapURL = "https://pro-api.coinmarketcap.com/v1/cryptocurrency/quotes/latest"
	FixerURL         = "https://data.fixer.io/api/latest"
)

func StartSubscription(ctx context.Context, logger *zap.SugaredLogger, cfg *config.Config) {

	d, err := New(cfg.DatabaseConnectionURL)
	if err != nil {
		logger.Fatal(err)
	}
	defer d.d.Close()

	var wg sync.WaitGroup
	for _, subscriber := range []func(context.Context, *sqlx.DB, *zap.SugaredLogger, *config.Config) error{
		SubscriptionBinance,
		SubscriptionCoinmarketcap,
		SubscriptionFixer,
		//...
	} {
		subscriber := subscriber
		wg.Add(1)
		go func() {
			defer wg.Done()
			SubscriptionWorker(ctx, d.d.DB, logger, cfg, subscriber)
		}()
	}

	wg.Wait()
}

func SubscriptionWorker(ctx context.Context, db *sqlx.DB, logger *zap.SugaredLogger, cfg *config.Config, fn func(context.Context, *sqlx.DB, *zap.SugaredLogger, *config.Config) error) {
	logger.Infow("INFO", "Database", "SubscriptionWorker Start")
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if err := fn(ctx, db, logger, cfg); err != nil {
			logger.Errorw("Database", "SubscriptionWorker", err)
		}

		interval, err := time.ParseDuration(cfg.Interval)
		if err != nil {
			logger.Errorw("Database", "SubscriptionWorker", err)
			return
		}
		time.Sleep(interval)
	}
}

func SubscriptionBinance(ctx context.Context, db *sqlx.DB, logger *zap.SugaredLogger, cfg *config.Config) error {
	client := http.Client{
		Timeout: 2 * time.Second,
	}
	Whitelisttokens, err := CnsTokenQuery(db)
	if err != nil {
		return fmt.Errorf("SubscriptionBinance CnsTokenQuery: %w", err)
	}
	if len(Whitelisttokens) == 0 {
		return fmt.Errorf("SubscriptionBinance CnsTokenQuery: The token does not exist.")
	}
	for _, token := range Whitelisttokens {
		tokensum := token + types.TokenBasecurrency

		req, err := http.NewRequest("GET", BinanceURL, nil)
		if err != nil {
			return fmt.Errorf("SubscriptionBinance fetch binance: %w", err)
		}
		q := url.Values{}
		q.Add("symbol", tokensum)
		req.Header.Set("Accepts", "application/json")
		req.URL.RawQuery = q.Encode()

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("SubscriptionBinance fetch binance: %w", err)
		}

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("SubscriptionBinance read body: %w", err)
		}
		if resp.StatusCode != 200 {
			bp := types.BinanceMsg{}
			err = json.Unmarshal(body, &bp)
			if err != nil {
				logger.Infow("SubscriptionBinance", resp.Status, "Request fail(apikey, symbol, rate-limited check)")
				return nil
			}
			logger.Infow("SubscriptionBinance", resp.Status, bp.Msg, "Request Symbol", token)
			return nil
		}
		bp := types.Binance{}
		err = json.Unmarshal(body, &bp)
		if err != nil {
			return fmt.Errorf("SubscriptionBinance unmarshal body: %w", err)
		}

		strToFloat, err := strconv.ParseFloat(bp.Price, 64)
		if strToFloat == float64(0) {
			continue
		}

		tx := db.MustBegin()
		now := time.Now()
		result := tx.MustExec("UPDATE oracle.binance SET price = ($1),updatedat = ($2) WHERE symbol = ($3)", strToFloat, now.Unix(), bp.Symbol)
		//https://www.cockroachlabs.com/docs/v20.2/alter-primary-key.html#alter-a-single-column-primary-key

		updateresult, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("SubscriptionBinance DB UPDATE: %w", err)
		}
		if updateresult == 0 {
			tx.MustExec("INSERT INTO oracle.binance VALUES (($1),($2),($3));", bp.Symbol, strToFloat, now.Unix())
		}

		err = tx.Commit()
		if err != nil {
			return fmt.Errorf("SubscriptionBinance DB commit: %w", err)
		}
		time.Sleep(1 * time.Second)
	}
	return nil
}

func SubscriptionCoinmarketcap(ctx context.Context, db *sqlx.DB, logger *zap.SugaredLogger, cfg *config.Config) error {
	client := &http.Client{
		Timeout: 2 * time.Second,
	}
	Whitelisttokens, err := CnsTokenQuery(db)
	if err != nil {
		return fmt.Errorf("SubscriptionCoinmarketcap CnsTokenQuery: %w", err)
	}
	if len(Whitelisttokens) == 0 {
		return fmt.Errorf("SubscriptionCoinmarketcap CnsTokenQuery: The token does not exist.")
	}
	req, err := http.NewRequest("GET", CoinmarketcapURL, nil)
	if err != nil {
		return fmt.Errorf("fetch coinmarketcap: %w", err)
	}
	q := url.Values{}
	q.Add("symbol", strings.Join(Whitelisttokens, ","))
	q.Add("convert", types.TokenBasecurrency)
	req.Header.Set("Accepts", "application/json")
	req.Header.Add("X-CMC_PRO_API_KEY", cfg.CoinmarketcapapiKey)
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("SubscriptionCoinmarketcap fetch coinmaketcap: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("SubscriptionCoinmarketcap read body: %w", err)
	}
	bp := types.Coinmarketcap{}
	if resp.StatusCode != 200 {
		err = json.Unmarshal(body, &bp)
		if err != nil {
			logger.Infow("SubscriptionCoinmarketcap", resp.Status, "Request fail(apikey, symbol, rate-limited check)")
			return nil
		}
		logger.Infow("SubscriptionCoinmarketcap", resp.Status, bp.Status.ErrorMessage, "Request Symbol", Whitelisttokens)
		return nil
	}
	var data map[string]struct {
		Quote struct {
			USDT struct {
				Price      float64 `json:"price"`
				Market_cap float64 `json:"market_cap"`
			} `json:"USDT"`
		} `json:"quote"`
	}

	err = json.Unmarshal(body, &bp)
	if err != nil {
		return fmt.Errorf("SubscriptionCoinmarketcap unmarshal body: %w", err)
	}
	err = json.Unmarshal(bp.Data, &data)
	if err != nil {
		return fmt.Errorf("SubscriptionCoinmarketcap unmarshal body: %w", err)
	}

	for _, token := range Whitelisttokens {
		tokensum := token + types.TokenBasecurrency
		d, ok := data[token]
		if !ok {
			return fmt.Errorf("SubscriptionCoinmarketcap price for symbol %q not found", tokensum)
		}

		tx := db.MustBegin()
		now := time.Now()

		resultsupply := tx.MustExec("UPDATE oracle.coinmarketcapsupply SET supply = ($1) WHERE symbol = ($2)", d.Quote.USDT.Market_cap, tokensum)

		updateresultsupply, err := resultsupply.RowsAffected()
		if err != nil {
			return fmt.Errorf("SubscriptionCoinmarketcap DB UPDATE: %w", err)
		}
		if updateresultsupply == 0 {
			tx.MustExec("INSERT INTO oracle.coinmarketcapsupply VALUES (($1),($2));", tokensum, d.Quote.USDT.Market_cap)
		}

		result := tx.MustExec("UPDATE oracle.coinmarketcap SET price = ($1),updatedat = ($2) WHERE symbol = ($3)", d.Quote.USDT.Price, now.Unix(), tokensum)

		updateresult, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("SubscriptionCoinmarketcap DB UPDATE: %w", err)
		}
		if updateresult == 0 {
			tx.MustExec("INSERT INTO oracle.coinmarketcap VALUES (($1),($2),($3));", tokensum, d.Quote.USDT.Price, now.Unix())
		}

		err = tx.Commit()
		if err != nil {
			return fmt.Errorf("SubscriptionCoinmarketcap DB commit: %w", err)
		}
		time.Sleep(1 * time.Second)
	}
	return nil
}

func SubscriptionFixer(ctx context.Context, db *sqlx.DB, logger *zap.SugaredLogger, cfg *config.Config) error {
	client := &http.Client{
		Timeout: 2 * time.Second,
	}
	req, err := http.NewRequest("GET", FixerURL, nil)
	if err != nil {
		return fmt.Errorf("SubscriptionFixer fetch Fixer: %w", err)
	}
	q := url.Values{}
	q.Add("access_key", cfg.Fixerapikey)
	q.Add("base", types.FiatBasecurrency)
	q.Add("symbols", strings.Join(cfg.Whitelistfiats, ","))

	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("SubscriptionFixer fetch Fixer: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("SubscriptionFixer read body: %w", err)
	}

	if resp.StatusCode != 200 {
		logger.Infow("SubscriptionFixer", resp.Status, "Request fail(apikey, symbol, rate-limited check)")
		return nil
	}

	bp := types.Fixer{}
	err = json.Unmarshal(body, &bp)
	if err != nil {
		return fmt.Errorf("SubscriptionFixer unmarshal body: %w", err)
	}
	if bp.Success != true {
		logger.Infow("SubscriptionFixer", bp.Success, "The status message of the query is fail(Maybe the apikey problem)")
		return nil
	}
	var data map[string]float64
	err = json.Unmarshal(bp.Rates, &data)
	if err != nil {
		return fmt.Errorf("SubscriptionFixer unmarshal body: %w", err)
	}

	for _, fiat := range cfg.Whitelistfiats {
		fiatsum := types.FiatBasecurrency + fiat
		d, ok := data[fiat]
		if !ok {
			logger.Infow("SubscriptionFixer", "From the provider list of deliveries price for symbol not found", fiatsum)
			return nil
		}

		tx := db.MustBegin()
		now := time.Now()
		result := tx.MustExec("UPDATE oracle.fixer SET price = ($1),updatedat = ($2) WHERE symbol = ($3)", d, now.Unix(), fiatsum)
		updateresult, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("SubscriptionFixer DB UPDATE: %w", err)
		}
		if updateresult == 0 {
			tx.MustExec("INSERT INTO oracle.fixer VALUES (($1),($2),($3));", fiatsum, d, now.Unix())
		}

		err = tx.Commit()
		if err != nil {
			return fmt.Errorf("SubscriptionFixer DB commit: %w", err)
		}
	}
	return nil
}
