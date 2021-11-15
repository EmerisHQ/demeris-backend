package rpcwatcher

import (
	"os"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/allinbits/demeris-backend/utils/store"
)

var (
	s  *store.Store
	mr *miniredis.Miniredis
)

func TestMain(m *testing.M) {
	mr, s = store.SetupTestDB()
	code := m.Run()
	defer mr.Close()
	os.Exit(code)
}

func TestHandleMessage(t *testing.T) {
	defer store.ResetTestDB(mr, s)
}
