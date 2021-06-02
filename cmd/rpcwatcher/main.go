package main

import (
	"context"
	"time"

	"github.com/allinbits/demeris-backend/models"
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

	var chains []models.Chain

	var cancelFunc context.CancelFunc

	var watchers []*rpcwatcher.Watcher

	err = db.Exec("select * from chains", map[string]interface{}{}, &chains)

	if err != nil {
		panic(err)
	}

	for _, c := range chains {
		watcher, err := rpcwatcher.NewWatcher(c.NodeInfo.Endpoint, l, db, s, []string{"tm.event='Tx'"})

		if err != nil {
			panic(err)
		}

		watchers = append(watchers, watcher)
	}

	cancelFunc = rpcwatcher.Start(watchers)

	for {
		var ch []models.Chain
		err = db.Exec("select * from chains", map[string]interface{}{}, &ch)

		if err != nil {
			panic(err)
		}

		if len(ch) != len(chains) {

			l.Infow("Chains modified. Restarting watchers")

			cancelFunc()

			chains = ch

			for _, c := range chains {
				watcher, err := rpcwatcher.NewWatcher(c.NodeInfo.Endpoint, l, db, s, []string{"tm.event='Tx'"})

				if err != nil {
					panic(err)
				}

				watchers = append(watchers, watcher)
			}

			cancelFunc = rpcwatcher.Start(watchers)

			time.Sleep(1 * time.Second)
		}
	}
}
