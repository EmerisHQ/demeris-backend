package database_test

import (
	"github.com/allinbits/demeris-backend/price-oracle/database"
	"github.com/cockroachdb/cockroach-go/v2/testserver"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	testServer := setup(t)
	defer tearDown(testServer)

	connStr := testServer.PGURL().String()
	require.NotNil(t, connStr)

	instance, err := database.New(connStr)
	require.NoError(t, err)
	require.Equal(t, instance.GetConnectionString(), connStr)

	rows, err := instance.Query("SHOW TABLES FROM oracle")
	require.NotNil(t, rows)

	var tableCountDB int
	for rows.Next() {
		tableCountDB++
	}
	err = rows.Err()
	require.NoError(t, err)

	err = rows.Close()
	require.NoError(t, err)

	var tableCountMigration int
	for _, migrationQuery := range database.MigrationList {
		if strings.HasPrefix(strings.TrimPrefix(migrationQuery, "\n"), "CREATE TABLE") {
			tableCountMigration++
		}
	}
	require.Equal(t, tableCountMigration, tableCountDB)
}

func TestCnstokenQueryHandler(t *testing.T) {
	testServer := setup(t)
	defer tearDown(testServer)

	instance, err := database.New(testServer.PGURL().String())
	require.NoError(t, err)

	_, err = instance.CnstokenQueryHandler()
	require.Error(t, err)
}

func setup(t *testing.T) testserver.TestServer {
	ts, err := testserver.NewTestServer()
	require.NoError(t, err)
	require.NoError(t, ts.WaitForInit())

	return ts
}

func tearDown(ts testserver.TestServer) {
	ts.Stop()
}
