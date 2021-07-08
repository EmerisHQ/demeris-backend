package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/r3labs/diff"

	apiconfig "github.com/allinbits/demeris-backend/api/config"

	apidb "github.com/allinbits/demeris-backend/api/database"
	cnsdb "github.com/allinbits/demeris-backend/cns/database"
	dbutils "github.com/allinbits/demeris-backend/utils/database"

	"github.com/allinbits/demeris-backend/models"
	"github.com/allinbits/demeris-backend/rpcwatcher"
	"github.com/allinbits/demeris-backend/utils/logging"
	"github.com/allinbits/demeris-backend/utils/store"
)

var Version = "not specified"

type watcherInstance struct {
	watcher *rpcwatcher.Watcher
	cancel  context.CancelFunc
}

func main() {
	c, err := rpcwatcher.ReadConfig()
	if err != nil {
		panic(err)
	}

	l := logging.New(logging.LoggingConfig{
		Debug: c.Debug,
	})

	l.Infow("rpcwatcher", "version", Version)

	tldb, err := apidb.Init(&apiconfig.Config{DatabaseConnectionURL: c.DatabaseConnectionURL})
	if err != nil {
		panic(err)
	}

	db, err := dbutils.New(c.DatabaseConnectionURL)
	if err != nil {
		panic(err)
	}

	cns, err := cnsdb.New(c.DatabaseConnectionURL)
	if err != nil {
		panic(err)
	}

	s := store.NewClient(c.RedisURL)

	var chains []models.Chain

	watchers := map[string]watcherInstance{}

<<<<<<< HEAD
	err = db.Exec("select * from cns.chains where enabled=TRUE", nil, &chains)
=======
	chains, err = cns.Chains()
>>>>>>> b5fc02d (rpcwatcher: use cns db instance)

	if err != nil {
		panic(err)
	}

	chainsMap := mapChains(chains)

	for cn := range chainsMap {
		watcher, err := rpcwatcher.NewWatcher(endpoint(cn), cn, l, db, tldb, cns, s, []string{"tm.event='Tx'"})

		if err != nil {
			l.Errorw("cannot create chain", "error", err)
			continue
		}

		ctx, cancel := context.WithCancel(context.Background())
		rpcwatcher.Start(watcher, ctx)

		watchers[cn] = watcherInstance{
			watcher: watcher,
			cancel:  cancel,
		}
	}

	for range time.Tick(1 * time.Second) {
<<<<<<< HEAD
		var ch []models.Chain
		err = db.Exec("select * from cns.chains where enabled=TRUE", nil, &ch)
=======
>>>>>>> b5fc02d (rpcwatcher: use cns db instance)

		ch, err := cns.Chains()

		newChainsMap := mapChains(ch)

		chainsDiff, err := diff.Diff(chainsMap, newChainsMap)
		if err != nil {
			l.Errorw("cannot diff maps", "error", err)
			continue
		}

		if chainsDiff == nil {
			continue
		}

		l.Infow("Chains modified. Restarting watchers")

		l.Debugw("diff", "diff", chainsDiff)
		for _, d := range chainsDiff {
			switch d.Type {
			case diff.DELETE:
				name := d.Path[0]
				wi, ok := watchers[name]
				if !ok {
					// we probably deleted this already somehow
					continue
				}
				wi.cancel()

				delete(watchers, name)
				delete(chainsMap, name)
			case diff.CREATE:
				name := d.Path[0]
				watcher, err := rpcwatcher.NewWatcher(endpoint(name), name, l, db, tldb, cns, s, []string{"tm.event='Tx'"})

				if err != nil {
					var dnsErr *net.DNSError
					if errors.As(err, &dnsErr) || strings.Contains(err.Error(), "connection refused") {
						l.Infow("chain not yet available", "name", name)
						continue
					}

					l.Errorw("cannot create chain", "error", err)
					continue
				}

				ctx, cancel := context.WithCancel(context.Background())
				rpcwatcher.Start(watcher, ctx)

				watchers[name] = watcherInstance{
					watcher: watcher,
					cancel:  cancel,
				}

				chainsMap[name] = newChainsMap[name]
			}
		}
	}
}

func mapChains(c []models.Chain) map[string]models.Chain {
	ret := map[string]models.Chain{}
	for _, cc := range c {
		ret[cc.ChainName] = cc
	}

	return ret
}

func endpoint(chainName string) string {
	return fmt.Sprintf("http://%s:26657", chainName)
}
