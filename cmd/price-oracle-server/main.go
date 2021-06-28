package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/allinbits/demeris-backend/price-oracle/config"
	"github.com/allinbits/demeris-backend/price-oracle/database"
	"github.com/allinbits/demeris-backend/price-oracle/rest"
	"github.com/allinbits/demeris-backend/utils/logging"
)

var Version = "not specified"

func main() {
	config, err := config.Read()
	if err != nil {
		panic(err)
	}

	logger := logging.New(logging.LoggingConfig{
		LogPath: config.LogPath,
		Debug:   config.Debug,
	})

	logger.Infow("price-oracle-server", "version", Version)

	di, err := database.New(config.DatabaseConnectionURL)
	if err != nil {
		logger.Fatal(err)
	}

	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg.Add(2)
	go func() {
		defer wg.Done()
		database.StartAggregate(ctx, logger, config)
	}()
	go func() {
		defer wg.Done()
		database.StartSubscription(ctx, logger, config)
	}()

	restServer := rest.NewServer(
		logger,
		di,
		config,
	)
	go func() {
		if err := restServer.Serve(config.ListenAddr); err != nil {
			logger.Panicw("rest http server error", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	cancel()
	wg.Wait()
	logger.Info("Shutting down server...")
}
