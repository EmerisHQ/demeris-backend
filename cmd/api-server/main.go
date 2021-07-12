package main

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"

	"github.com/allinbits/demeris-backend/api/config"
	"github.com/allinbits/demeris-backend/api/database"
	"github.com/allinbits/demeris-backend/api/router"
	"github.com/allinbits/demeris-backend/utils/k8s"
	"github.com/allinbits/demeris-backend/utils/logging"
	"github.com/allinbits/demeris-backend/utils/store"
	gaia "github.com/cosmos/gaia/v4/app"
)

var Version = "not specified"

const trim    = "__keyspace@0__:shadow"

func main() {
	cfg, err := config.Read()
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

	l.Infow("api-server", "version", Version)

	dbi, err := database.Init(cfg)
	if err != nil {
		l.Panicw("cannot initialize database", "error", err)
	}

	s := store.NewClient(cfg.RedisAddr)
	s.Client.ConfigSet(s.Client.Context(), "notify-keyspace-events", "Kx")

	sub := s.Client.PSubscribe(s.Client.Context(), "__key*__:*")
		for msg := range sub.Channel() {
			l.Infow("new message received","msg",  msg.Channel)
			k := strings.TrimPrefix(msg.Channel, trim)
			l.Infow("this is string", "string", k )
			if s.Exists(k){
				ticket, err := s.Get(k)
				l.Infow("this is value", "ticket", ticket, "err", err)
				if err != nil {
					panic(err)
				}
				ticket.Status = fmt.Sprintf("stuck_%s", ticket.Status)
				if err := s.Set(k, ticket, 0); err != nil {
					panic(err)
				}
			}
		}

	kubeClient, err := k8s.NewInCluster()
	if err != nil {
		l.Panicw("cannot initialize k8s", "error", err)
	}

	cdc, _ := gaia.MakeCodecs()

	r := router.New(
		dbi,
		l,
		s,
		kubeClient,
		cfg.KubernetesNamespace,
		cfg.CNSAddr,
		cdc,
		cfg.Debug,
	)

	if err := r.Serve(cfg.ListenAddr); err != nil {
		l.Panicw("http server panic", "error", err)
	}
}
