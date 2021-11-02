package database_test

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/allinbits/demeris-backend/price-oracle/config"
	"github.com/allinbits/demeris-backend/price-oracle/database"
	"github.com/allinbits/demeris-backend/price-oracle/types"
	dbutils "github.com/allinbits/demeris-backend/utils/database"
	"github.com/allinbits/demeris-backend/utils/logging"
	"github.com/cockroachdb/cockroach-go/v2/testserver"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestStartSubscription(t *testing.T) {
	binance := types.Binance{
		Symbol: "ATOMUSDT",
		Price:  "-50.0", // A value that is never possible in real world.
	}
	b, err := json.Marshal(binance)
	require.NoError(t, err)

	ctx, cancel, logger, cfg, tDown := setupSubscription(t)
	defer tDown()
	defer cancel()

	instance, err := database.New(cfg.DatabaseConnectionURL)
	require.NoError(t, err)

	client := newTestClient(func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader(b)),
		}
	})
	api := database.Api{
		Client:   client,
		Instance: instance,
	}

	err = api.SubscriptionBinance(ctx, logger, cfg)
	require.NoError(t, err)

	price := getTokenPrices(t, cfg.DatabaseConnectionURL, "oracle.binance", []string{"ATOMUSDT"})
	require.Equal(t, price["ATOMUSDT"], -50.0)
}

type roundTripFunc func(req *http.Request) *http.Response

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

// newTestClient is with our transport.
func newTestClient(fn roundTripFunc) *http.Client {
	return &http.Client{
		Transport: fn,
	}
}

func setupSubscription(t *testing.T) (context.Context, func(), *zap.SugaredLogger, *config.Config, func()) {
	t.Helper()
	testServer, err := testserver.NewTestServer()
	require.NoError(t, err)
	require.NoError(t, testServer.WaitForInit())

	connStr := testServer.PGURL().String()
	require.NotNil(t, connStr)

	// Seed DB with data in schema file
	oracleMigration := readLinesFromFile(t, "schema-unittest")
	err = dbutils.RunMigrations(connStr, oracleMigration)
	require.NoError(t, err)

	cfg := &config.Config{ // config.Read() is not working. Fixing is not in scope of this task. That comes later.
		LogPath:               "",
		Debug:                 true,
		DatabaseConnectionURL: connStr,
		Interval:              "10s",
		Whitelistfiats:        []string{"EUR", "KRW", "CHF"},
	}

	logger := logging.New(logging.LoggingConfig{
		LogPath: cfg.LogPath,
		Debug:   cfg.Debug,
	})

	insertToken(t, connStr)
	ctx, cancel := context.WithCancel(context.Background())
	return ctx, cancel, logger, cfg, func() { testServer.Stop() }
}

func getTokenPrices(t *testing.T, connStr, tableName string, symbols []string) map[string]float64 {
	instance, err := database.New(connStr)
	require.NoError(t, err)

	tokenPrice := make(map[string]float64)
	rows, err := instance.Query("SELECT * FROM " + tableName)
	require.NoError(t, err)
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var tokenName string
		var price float64
		var updatedAt float64
		err := rows.Scan(&tokenName, &price, &updatedAt)
		require.NoError(t, err)
		tokenPrice[tokenName] = price
	}
	ret := make(map[string]float64)
	for _, symbol := range symbols {
		ret[symbol] = tokenPrice[symbol]
	}
	return ret
}
