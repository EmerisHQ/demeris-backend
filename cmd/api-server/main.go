package main

import (
	"log"
	"net/http"
	"runtime"

	"github.com/allinbits/demeris-backend/api/config"
	"github.com/allinbits/demeris-backend/api/database"
	"github.com/allinbits/demeris-backend/api/router"
	"github.com/allinbits/demeris-backend/utils/k8s"
	"github.com/allinbits/demeris-backend/utils/logging"
	"github.com/allinbits/demeris-backend/utils/store"
	gaia "github.com/cosmos/gaia/v4/app"
)

var Version = "not specified"

func main() {
	cfg, err := config.Read()
	if err != nil {
		panic(err)
	}

	if cfg.Debug {
		runtime.SetCPUProfileRate(500)

		go func() {
			log.Printf("Starting Server! \t Go to http://localhost:6060/debug/pprof/\n")
			err := http.ListenAndServe("localhost:6060", nil)
			if err != nil {
				panic(err)
			}
		}()
	}

	l := logging.New(logging.LoggingConfig{
		Debug: cfg.Debug,
	})

	l.Infow("api-server", "version", Version)

	dbi, err := database.Init(cfg)
	if err != nil {
		l.Panicw("cannot initialize database", "error", err)
	}

	s := store.NewClient(cfg.RedisAddr)

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
