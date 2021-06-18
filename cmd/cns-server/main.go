package main

import (
	"github.com/allinbits/demeris-backend/cns/chainwatch"
	"github.com/allinbits/demeris-backend/cns/database"
	"github.com/allinbits/demeris-backend/cns/rest"
	"github.com/allinbits/demeris-backend/utils/k8s"
	"github.com/allinbits/demeris-backend/utils/logging"
)

func main() {
	config, err := readConfig()
	if err != nil {
		panic(err)
	}

	logger := logging.New(logging.LoggingConfig{
		LogPath: config.LogPath,
		Debug:   config.Debug,
	})

	di, err := database.New(config.DatabaseConnectionURL)
	if err != nil {
		logger.Fatal(err)
	}

	kube, err := k8s.NewInCluster()
	if err != nil {
		logger.Fatal(err)
	}

	rc, err := chainwatch.NewConnection(config.Redis)
	if err != nil {
		logger.Fatal(err)
	}

	ci := chainwatch.New(
		logger,
		kube,
		rc,
		di,
	)

	go ci.Run()

	restServer := rest.NewServer(
		logger,
		di,
		&kube,
		rc,
		config.Debug,
	)

	if err := restServer.Serve(config.RESTAddress); err != nil {
		logger.Panicw("rest http server error", "error", err)
	}
}
