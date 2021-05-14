package main

import (
	"github.com/allinbits/demeris-backend/utils/k8s"
	"github.com/allinbits/demeris-backend/utils/logging"

	"github.com/allinbits/demeris-backend/api/config"
	"github.com/allinbits/demeris-backend/api/database"
	"github.com/allinbits/demeris-backend/api/router"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		panic(err)
	}

	l := logging.New(logging.LoggingConfig{
		Debug: cfg.Debug,
	})

	dbi, err := database.Init(cfg)
	if err != nil {
		l.Panicw("cannot initialize database", "error", err)
	}

	kubeClient, err := k8s.New()
	if err != nil {
		l.Panicw("cannot initialize k8s", "error", err)
	}

	r := router.New(
		dbi,
		l,
		kubeClient,
		cfg.CNSAddr,
	)

	if err := r.Serve(cfg.ListenAddr); err != nil {
		l.Panicw("http server panic", "error", err)
	}
}
