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
	"github.com/allinbits/demeris-backend/rpcwatcher/database"
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

	s, err := store.NewClient(c.RedisURL)
	if err != nil {
		l.Panicw("unable to start redis client", "error", err)
	}
	var chains []models.Chain

	watchers := map[string]watcherInstance{}

	chains, err = db.Chains()

	if err != nil {
		panic(err)
	}

	chainsMap := mapChains(chains)

	for cn := range chainsMap {
		subEvents := []string{rpcwatcher.EventsTx, rpcwatcher.EventsBlock}

		watcher, err := rpcwatcher.NewWatcher(endpoint(cn), cn, l, c.ApiURL, db, s, subEvents)

		if err != nil {
			l.Errorw("cannot create chain", "error", err)
			delete(chainsMap, cn)
			continue
		}

		err = s.SetWithExpiry(cn, "true", 0)
		if err != nil {
			l.Errorw("unable to set chain name as true", "error", err)
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

		ch, err := db.Chains()

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

				subEvents := []string{rpcwatcher.EventsTx}

				if name == "cosmos-hub" { // special case, needs to observe new blocks too
					subEvents = append(subEvents, rpcwatcher.EventsBlock)
				}

				watcher, err := rpcwatcher.NewWatcher(endpoint(name), name, l, c.ApiURL, db, s, subEvents)

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
				err = s.SetWithExpiry(name, "true", 0)
				if err != nil {
					l.Errorw("unable to set chain name as true", "error", err)
				}

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
