package database_test

import (
	"bufio"
	"context"
	cnsDB "github.com/allinbits/demeris-backend/cns/database"
	"github.com/allinbits/demeris-backend/models"
	"github.com/allinbits/demeris-backend/price-oracle/config"
	"github.com/allinbits/demeris-backend/price-oracle/database"
	dbutils "github.com/allinbits/demeris-backend/utils/database"
	"github.com/allinbits/demeris-backend/utils/logging"
	"github.com/cockroachdb/cockroach-go/v2/testserver"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(logger)
	os.Exit(m.Run())
}

func TestStartAggregate(t *testing.T) {
	ctx, cancel, logger, cfg, tDown := setupAgg(t)
	defer tDown()
	defer cancel()

	atomPrice, lunaPrice := getTokenPrice(t, cfg.DatabaseConnectionURL)
	require.Equal(t, atomPrice, 10.0)
	require.Equal(t, lunaPrice, 10.0)

	go database.StartAggregate(ctx, logger, cfg)
	// Validate data updated on DB ..
	time.Sleep(5 * time.Second)

	atomPrice, lunaPrice = getTokenPrice(t, cfg.DatabaseConnectionURL)
	require.Equal(t, atomPrice, 15.0)
	require.Equal(t, lunaPrice, 16.0)
}

func setupAgg(t *testing.T) (context.Context, func(), *zap.SugaredLogger, *config.Config, func()) {
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

	ctx, cancel := context.WithCancel(context.Background())
	return ctx, cancel, logger, cfg, func() { testServer.Stop() }
}

func getTokenPrice(t *testing.T, connStr string) (float64, float64) {
	instance, err := database.New(connStr)
	require.NoError(t, err)

	tokenPrice := make(map[string]float64)
	rows, err := instance.Query("SELECT * FROM oracle.tokens")
	require.NoError(t, err)
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var tokenName string
		var price float64
		err := rows.Scan(&tokenName, &price)
		require.NoError(t, err)
		tokenPrice[tokenName] = price
	}
	return tokenPrice["ATOMUSDT"], tokenPrice["LUNAUSDT"]
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
