package main

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/allinbits/demeris-backend/price-oracle/config"
	"github.com/allinbits/demeris-backend/price-oracle/db"
	"github.com/allinbits/demeris-backend/price-oracle/server"
)

func InitZapLog() *zap.Logger {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	var err error
	logger, err := config.Build()
	if err != nil {
		logger.Panic("Panic",
			zap.String("Zap Logger", err.Error()),
			zap.Duration("Duration", time.Second),
		)
	}
	return logger
}

func main() {
	logger := InitZapLog()
	defer logger.Sync()

	config.ReadConfig(logger)

	logger.Info("INFO",
		zap.String("Oracle", "Start oracle"),
		zap.Duration("Duration", time.Second),
	)

	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wg.Add(1)
	go db.StartAggregate(&wg, ctx, logger)
	wg.Add(1)
	go db.StartSubscription(&wg, ctx, logger)
	wg.Add(1)
	go server.StartServer(&wg, ctx, logger)

	wg.Wait()
	logger.Fatal("Fatal",
		zap.String("Oracle", "Stop oracle"),
		zap.Duration("Duration", time.Second),
	)
}
