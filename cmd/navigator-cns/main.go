package main

import (
	"github.com/allinbits/navigator-cns/rest"
	"github.com/allinbits/navigator-utils/logging"

	"github.com/allinbits/navigator-cns/database"
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
