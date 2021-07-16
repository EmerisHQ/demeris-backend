package main

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"

	"github.com/allinbits/demeris-backend/utils/logging"
	"github.com/allinbits/demeris-backend/utils/store"
)

var Version = "not specified"

const prefix = "__keyspace@0__:shadow"

func main() {
	cfg, err := readConfig()
	if err != nil {
		panic(err)
	}

	l := logging.New(logging.LoggingConfig{
		Debug: cfg.Debug,
	})

	if cfg.Debug {
		runtime.SetCPUProfileRate(500)

		go func() {
			l.Debugw("starting profiling server", "port", "6060")
			err := http.ListenAndServe(":6060", nil)
			if err != nil {
				l.Panicw("cannot run profiling server", "error", err)
			}
		}()
	}

	l.Infow("ticket-watcher", "version", Version)
	s := store.NewClient(cfg.RedisUrl)
	s.Client.ConfigSet(s.Client.Context(), "notify-keyspace-events", "Kx")

	sub := s.Client.PSubscribe(s.Client.Context(), "__key*__:*")
	for msg := range sub.Channel() {
		l.Debugw("new message received", "msg", msg.Channel)
		key := strings.TrimPrefix(msg.Channel, prefix)
		l.Debugw("Received key after trim", "key", key)

		if s.Exists(key) {

			ticket, err := s.Get(key)
			if err != nil {
				l.Errorw("unable to get ticket value to get error", "error", err)
				continue
			}

			ticket.Status = fmt.Sprintf("stuck_%s", ticket.Status)
			if err := s.SetWithExpiry(key, ticket, 0); err != nil {
				l.Errorw("unable to set ticket value to stuck", "error", err)
			}
		}
	}
}
