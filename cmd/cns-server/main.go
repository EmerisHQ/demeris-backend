package main

import (
	"github.com/allinbits/demeris-backend/cns/database"
	"github.com/allinbits/demeris-backend/cns/rest"
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

	restServer := rest.NewServer(
		logger,
		di,
		config.Debug,
	)

	if err := restServer.Serve(config.RESTAddress); err != nil {
		logger.Panicw("rest http server error", "error", err)
	}
}
