package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/allinbits/demeris-backend/price-oracle/config"
)

var Logger *zap.Logger

func StartServer(wg *sync.WaitGroup, ctx context.Context, logger *zap.Logger) {
	defer wg.Done()
	Logger = logger
	ConnectDB(wg)
	defer db.Close()

	router := gin.New()
	gin.SetMode(gin.ReleaseMode)
	router.GET("/", AllTokenPrices)
	router.GET("/prices", AllTokenPrices)
	router.POST("/tokens", TokensPrices)
	router.POST("/fiat", FiatsPrices)

	client := &http.Server{
		Addr:    config.Config.Laddr,
		Handler: router,
	}

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	go func() {
		<-quit
		logger.Info("INFO",
			zap.String("Server", "receive interrupt signal/Close"),
			zap.Duration("Duration", time.Second),
		)
		wg.Done()
		if err := db.Close(); err != nil {
			logger.Fatal("Fatal",
				zap.String("Server", err.Error()),
				zap.Duration("Duration", time.Second),
			)
		}
		if err := client.Close(); err != nil {
			logger.Fatal("Fatal",
				zap.String("Server", err.Error()),
				zap.Duration("Duration", time.Second),
			)
		}
	}()

	if config.Config.SSLMode == true {
		if err := client.ListenAndServeTLS(config.Config.SSLCrt, config.Config.SSLKey); err != nil {
			if err == http.ErrServerClosed {
				logger.Info("INFO",
					zap.String("Server", "Server closed under request"),
					zap.Duration("Duration", time.Second),
				)
			} else {
				logger.Fatal("Fatal",
					zap.String("Server", err.Error()),
					zap.Duration("Duration", time.Second),
				)
			}
		}
	} else if config.Config.SSLMode == false {
		if err := client.ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
				logger.Info("INFO",
					zap.String("Server", "Server closed under request"),
					zap.Duration("Duration", time.Second),
				)
			} else {
				logger.Fatal("Fatal",
					zap.String("Server", err.Error()),
					zap.Duration("Duration", time.Second),
				)
			}
		}
	}
	logger.Info("INFO",
		zap.String("Server", "Server exiting"),
		zap.Duration("Duration", time.Second),
	)
}
