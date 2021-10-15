package rest

import (
	"bufio"
	"encoding/json"
	"github.com/alicebob/miniredis/v2"
	cnsDB "github.com/allinbits/demeris-backend/cns/database"
	"github.com/allinbits/demeris-backend/models"
	"github.com/allinbits/demeris-backend/price-oracle/config"
	"github.com/allinbits/demeris-backend/price-oracle/database"
	"github.com/allinbits/demeris-backend/price-oracle/types"
	dbutils "github.com/allinbits/demeris-backend/utils/database"
	"github.com/allinbits/demeris-backend/utils/logging"
	"github.com/allinbits/demeris-backend/utils/store"
	"github.com/cockroachdb/cockroach-go/v2/testserver"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestAllPricesHandler(t *testing.T) {
	router, ctx, w, tDown := setup(t)
	defer tDown()
	wantFiats := []types.FiatPriceResponse{
		{Symbol: "USDCHF", Price: 10},
		{Symbol: "USDEUR", Price: 20},
		{Symbol: "USDKRW", Price: 5},
	}
	wantToken := []types.TokenPriceResponse{
		{Price: 10, Symbol: "ATOMUSDT", Supply: 113563929433.0},
		{Price: 10, Symbol: "LUNAUSDT", Supply: 113563929433.0},
	}
	router.allPricesHandler(ctx)
	gotFiats := getFiatsFromResponseWriter(t, w)
	require.Equal(t, wantFiats, gotFiats)
	gotToken := getCryptoFromResponseWriter(t, w)
	require.Equal(t, wantToken, gotToken)
}

func setup(t *testing.T) (router, *gin.Context, *httptest.ResponseRecorder, func()) {
	t.Helper()
	// Setup DB
	tServer, err := testserver.NewTestServer()
	require.NoError(t, err)

	require.NoError(t, tServer.WaitForInit())

	connStr := tServer.PGURL().String()
	require.NotNil(t, connStr)

	// Seed DB with data in schema file
	oracleMigration := readLinesFromFile(t, "../database/schema-unittest")
	err = dbutils.RunMigrations(connStr, oracleMigration)
	require.NoError(t, err)
	// Put dummy data in cns DB
	insertToken(t, connStr)

	// Setup redis
	minRedis, err := miniredis.Run()
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

	dbInstance, err := database.New(connStr)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	ctx, engine := gin.CreateTestContext(w)

	str, err := store.NewClient(minRedis.Addr())
	require.NoError(t, err)

	server := &Server{
		l:  logger,
		d:  dbInstance,
		c:  cfg,
		g:  engine,
		ri: str,
	}

	return router{s: server}, ctx, w, func() { tServer.Stop(); minRedis.Close() }
}

func insertToken(t *testing.T, connStr string) {
	chain := models.Chain{
		ChainName:        "cosmos-hub",
		DemerisAddresses: []string{"addr1"},
		DisplayName:      "ATOM display name",
		GenesisHash:      "hash",
		NodeInfo:         models.NodeInfo{},
		ValidBlockThresh: models.Threshold(1 * time.Second),
		DerivationPath:   "derivation_path",
		SupportedWallets: []string{"metamask"},
		Logo:             "logo 1",
		Denoms: []models.Denom{
			{
				Name:        "ATOM",
				DisplayName: "ATOM",
				FetchPrice:  true,
				Ticker:      "ATOM",
			},
			{
				Name:        "LUNA",
				DisplayName: "LUNA",
				FetchPrice:  true,
				Ticker:      "LUNA",
			},
		},
		PrimaryChannel: models.DbStringMap{
			"cosmos-hub":  "ch0",
			"persistence": "ch2",
		},
	}
	cnsInstanceDB, err := cnsDB.New(connStr)
	require.NoError(t, err)

	err = cnsInstanceDB.AddChain(chain)
	require.NoError(t, err)

	cc, err := cnsInstanceDB.Chains()
	require.NoError(t, err)
	require.Equal(t, 1, len(cc))
}

func getFiatsFromResponseWriter(t *testing.T, w *httptest.ResponseRecorder) []types.FiatPriceResponse {
	var h gin.H
	err := json.Unmarshal(w.Body.Bytes(), &h)
	require.NoError(t, err)

	x := h["data"].(map[string]interface{})
	fiats := make([]types.FiatPriceResponse, 0)

	for _, ff := range x["Fiats"].([]interface{}) {
		fiats = append(fiats, types.FiatPriceResponse{
			Price:  ff.(map[string]interface{})["Price"].(float64),
			Symbol: ff.(map[string]interface{})["Symbol"].(string),
		})
	}
	return fiats
}

func getCryptoFromResponseWriter(t *testing.T, w *httptest.ResponseRecorder) []types.TokenPriceResponse {
	var h gin.H
	err := json.Unmarshal(w.Body.Bytes(), &h)
	require.NoError(t, err)

	x := h["data"].(map[string]interface{})
	tokens := make([]types.TokenPriceResponse, 0)

	for _, ff := range x["Tokens"].([]interface{}) {
		tokens = append(tokens, types.TokenPriceResponse{
			Price:  ff.(map[string]interface{})["Price"].(float64),
			Supply: ff.(map[string]interface{})["Supply"].(float64),
			Symbol: ff.(map[string]interface{})["Symbol"].(string),
		})
	}
	return tokens
}

func readLinesFromFile(t *testing.T, s string) []string {
	file, err := os.Open(s)
	require.NoError(t, err)
	defer func() { _ = file.Close() }()

	var commands []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		cmd := scanner.Text()
		commands = append(commands, cmd)
	}
	return commands
}