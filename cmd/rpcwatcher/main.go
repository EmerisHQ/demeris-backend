package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/r3labs/diff"

	"github.com/allinbits/demeris-backend/models"
	"github.com/allinbits/demeris-backend/rpcwatcher"
	"github.com/allinbits/demeris-backend/utils/database"
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

	db, err := database.New(c.DatabaseConnectionURL)
	if err != nil {
		panic(err)
	}

	s := store.NewClient(c.RedisURL)

	var chains []models.Chain

	watchers := map[string]watcherInstance{}

	err = db.Exec("select * from cns.chains where enabled=TRUE", nil, &chains)

	if err != nil {
		panic(err)
	}

	chainsMap := mapChains(chains)

	for cn := range chainsMap {
		l.Debugw("connecting to chain", "chainName", cn)
		watcher, err := rpcwatcher.NewWatcher(endpoint(cn), cn, l, db, s, []string{"tm.event='Tx'"})

		if err != nil {
			l.Errorw("cannot create chain", "error", err)
			delete(chainsMap, cn)
			continue
		}

		l.Debugw("connected", "chainName", cn)

		ctx, cancel := context.WithCancel(context.Background())
		rpcwatcher.Start(watcher, ctx)

		watchers[cn] = watcherInstance{
			watcher: watcher,
			cancel:  cancel,
		}
	}

	for range time.Tick(1 * time.Second) {
		var ch []models.Chain
		err = db.Exec("select * from cns.chains where enabled=TRUE", nil, &ch)

		if err != nil {
			panic(err)
		}

		newChainsMap := mapChains(ch)

		chainsDiff, err := diff.Diff(chainsMap, newChainsMap)
		if err != nil {
			l.Errorw("cannot diff maps", "error", err)
			continue
		}

		if chainsDiff == nil {
			continue
		}

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
				watcher, err := rpcwatcher.NewWatcher(endpoint(name), name, l, db, s, []string{"tm.event='Tx'"})

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
