package lightwatcher

import (
	"github.com/tendermint/tendermint/libs/log"
	"go.uber.org/zap"
)

type zapLogger struct {
	z         *zap.SugaredLogger
	chainName string
}

func getLogFields(chainName string, keyVals ...interface{}) []interface{} {
	fields := []interface{}{
		"chain_name",
		chainName,
	}

	if len(keyVals)%2 != 0 {
		return fields
	}

	for i := 0; i < len(keyVals); i += 2 {
		fields = append(fields, keyVals[i].(string))
		fields = append(fields, keyVals[i+1])
	}

	return fields
}

func (z zapLogger) Debug(msg string, keyvals ...interface{}) {
	z.z.Debugw(msg, getLogFields(z.chainName, keyvals)...)
}

func (z zapLogger) Info(msg string, keyvals ...interface{}) {
	z.z.Infow(msg, getLogFields(z.chainName, keyvals)...)
}

func (z zapLogger) Error(msg string, keyvals ...interface{}) {
	z.z.Errorw(msg, getLogFields(z.chainName, keyvals)...)
}

func (z zapLogger) With(_ ...interface{}) log.Logger {
	return z
}
