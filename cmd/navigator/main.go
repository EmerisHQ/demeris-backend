package main

import (
	"github.com/allinbits/navigator-utils/logging"

	"github.com/allinbits/navigator-backend/config"
	"github.com/allinbits/navigator-backend/database"
	"github.com/allinbits/navigator-backend/router"
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
		l.Panicw("cannot initialize database", err)
	}

	r := router.New(
		dbi,
		l,
	)

	if err := r.Serve(cfg.ListenAddr); err != nil {
		l.Panicw("http server panic", err)
	}
}
