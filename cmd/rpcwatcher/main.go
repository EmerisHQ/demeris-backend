package main

import (
	"github.com/allinbits/demeris-backend/rpcwatcher"
	"github.com/allinbits/demeris-backend/utils/database"
	"github.com/allinbits/demeris-backend/utils/logging"
	"github.com/allinbits/demeris-backend/utils/store"
)

func main() {
	c, err := rpcwatcher.ReadConfig()
	if err != nil {
		panic(err)
	}

	l := logging.New(logging.LoggingConfig{
		Debug: c.Debug,
	})

	db, err := database.New(c.DatabaseConnectionURL)
	if err != nil {
		panic(err)
	}

	s := store.NewClient(c.RedisURL)

	watcher, err := rpcwatcher.NewWatcher(c.TendermintNode, l, db, s, []string{"tm.event='Tx'"})

	for data := range watcher.DataChannel {
		watcher.HandleMessage(data)
	}
}
