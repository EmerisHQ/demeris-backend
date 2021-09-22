package main

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof"

	"github.com/allinbits/demeris-backend/lightwatcher"
	"github.com/allinbits/demeris-backend/utils/logging"
)

var Version = "not specified"

var (
	eventsToSubTo = []string{lightwatcher.EventsTx, lightwatcher.EventsBlock}
)

func main() {
	c, err := lightwatcher.ReadConfig()
	if err != nil {
		panic(err)
	}

	l := logging.New(logging.LoggingConfig{
		Debug: true,
		JSON:  true,
	})

	l.Infow("lightwatcher", "version", Version)

	go func() {
		l.Debugw("starting profiling server", "port", "6061")
		err := http.ListenAndServe(":6061", nil)
		if err != nil {
			l.Panicw("cannot run profiling server", "error", err)
		}
	}()

	watcher, err := lightwatcher.NewWatcher(
		endpoint(c.ChainName),
		c.ChainName,
		l,
		eventsToSubTo,
	)

	if err != nil {
		l.Panicw("cannot create chain", "error", err)
	}

	l.Debugw("connected", "chainName", c.ChainName)
	ctx, _ := context.WithCancel(context.Background())
	lightwatcher.Start(watcher, ctx)

	<-make(chan struct{})
}

func endpoint(chainName string) string {
	return fmt.Sprintf("http://%s:26657", chainName)
}
