package main

import (
	"github.com/allinbits/demeris-backend/utils/logging"
	"github.com/allinbits/demeris-backend/utils/store"
	gaia "github.com/cosmos/gaia/v4/app"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/allinbits/demeris-backend/api/config"
	"github.com/allinbits/demeris-backend/api/database"
	"github.com/allinbits/demeris-backend/api/router"
)

var Version = "not specified"

func main() {
	cfg, err := config.Read()
	if err != nil {
		panic(err)
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

	/*kubeClient, err := k8s.NewInCluster()
	if err != nil {
		l.Panicw("cannot initialize k8s", "error", err)
	}*/

	var c client.Client
	cdc, _ := gaia.MakeCodecs()

	r := router.New(
		dbi,
		l,
		s,
		c,
		cfg.CNSAddr,
		cdc,
	)

	if err := r.Serve(cfg.ListenAddr); err != nil {
		l.Panicw("http server panic", "error", err)
	}
}
