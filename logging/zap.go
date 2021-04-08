package logging

import (
	"os"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LoggingConfig struct {
	LogPath string
	Debug   bool
}

// New creates a zap.SugaredLogger configured following lc's directives.
// If lc fields are empty, a debug logger will be returned.
// The logger returned by New is configured to automatically compress, rotate logs every 28 days or when
// they reach 20 megabytes.
func New(lc LoggingConfig) *zap.SugaredLogger {
	if lc == (LoggingConfig{}) {
		lc.Debug = true
	}

	if lc.Debug {
		// we can safely ignore the error here
		dc, _ := zap.NewDevelopment()
		return dc.Sugar()
	}

	var cores []zapcore.Core

	l := &lumberjack.Logger{
		Filename:   lc.LogPath,
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
