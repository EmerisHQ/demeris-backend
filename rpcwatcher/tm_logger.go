package rpcwatcher

import (
	"github.com/tendermint/tendermint/libs/log"
	"go.uber.org/zap"
)

type zapLogger struct {
	z         *zap.SugaredLogger
	chainName string
}

func (z zapLogger) Debug(msg string, keyvals ...interface{}) {
	z.z.Debugw(msg, "chain_name", z.chainName, keyvals)
}

func (z zapLogger) Info(msg string, keyvals ...interface{}) {
	z.z.Infow(msg, "chain_name", z.chainName, keyvals)
}

func (z zapLogger) Error(msg string, keyvals ...interface{}) {
	z.z.Errorw(msg, "chain_name", z.chainName, keyvals)
}

func (z zapLogger) With(_ ...interface{}) log.Logger {
	return z
}
