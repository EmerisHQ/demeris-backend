package store_test

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/allinbits/demeris-backend/utils/store"
	"github.com/stretchr/testify/require"
)

const (
	testChain  = "cosmos-hub"
	testTxHash = "918DC23785CABA3EE4E4A59321E679F8B7A2E27C9DFB165B3B6D22EF23017264"
	testOwner  = "cosmos1l2lepugxx5heetsl2cs74e2sy0uqxv390as7zw"
)

var s *store.Store

func TestMain(m *testing.M) {
	mr, err := miniredis.Run()
	if err != nil {
		log.Fatalf("got error: %s when running miniredis", err)
	}

	s, err = store.NewClient(mr.Addr())
	if err != nil {
		log.Fatalf("got error: %s when creating new store client", err)
	}

	code := m.Run()
	defer mr.Close()
	os.Exit(code)
}

func TestCreateTicket(t *testing.T) {
	require.NoError(t, s.CreateTicket(testChain, testTxHash, testOwner))
	key := fmt.Sprintf("%s/%s", testChain, testTxHash)
	require.True(t, s.Exists(key))
}
