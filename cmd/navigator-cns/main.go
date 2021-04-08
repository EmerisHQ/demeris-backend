package main

import (
	"os"

	"github.com/allinbits/navigator-cns/rest"

	"github.com/allinbits/navigator-cns/database"

	"go.uber.org/zap/zapcore"

	"go.uber.org/zap"

	"gopkg.in/natefinch/lumberjack.v2"
)

func main() {
	config, err := readConfig()
	if err != nil {
		panic(err)
	}

	logger := logging(config)

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

func logging(c *Config) *zap.SugaredLogger {
	if c.Debug {
		// we can safely ignore the error here
		dc, _ := zap.NewDevelopment()
		return dc.Sugar()
	}

	var cores []zapcore.Core

	l := &lumberjack.Logger{
		Filename:   c.LogPath,
		MaxSize:    20,
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   true,
	}

	fileLogger := zapcore.AddSync(l)
	jsonWriter := zapcore.AddSync(os.Stdout)

	cores = append(cores, zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		fileLogger,
		zap.InfoLevel,
	))

	// we use development encoder config in CLI output because it's easier to read
	cores = append(cores, zapcore.NewCore(
		zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
		jsonWriter,
		zap.InfoLevel,
	))

	logger := zap.New(zapcore.NewTee(cores...))

	return logger.WithOptions(zap.AddCaller()).Sugar()
}
