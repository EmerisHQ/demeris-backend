package tests

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	utils "github.com/allinbits/demeris-backend/test_utils"
	"github.com/allinbits/emeris-utils/logging"
	"go.uber.org/zap"
)

const baseUrl = "%s://%s%s"

var testCtx struct {
	emIngress utils.EmerisIngress
	client    *http.Client
	chains    []utils.EnvChain
}

func TestMain(m *testing.M) {
	logger := logging.New(logging.LoggingConfig{
		LogPath: "",
		Debug:   true,
	})

	env := os.Getenv("ENV")
	checkNotNil(env, "env", logger)

	emIngress, _, err := utils.LoadIngressInfo(env)
	checkNoError(err, logger)

	testCtx.emIngress = emIngress

	chains, err := utils.LoadChainsInfo(env)
	checkNoError(err, logger)

	testCtx.chains = chains

	client, err := utils.CreateNetClient(env)
	checkNoError(err, logger)

	testCtx.client = client

	exitVal := m.Run()

	os.Exit(exitVal)
}

func checkNoError(err error, logger *zap.SugaredLogger) {
	if err != nil {
		logger.Error(err)
		os.Exit(-1)
	}
}

func checkNotNil(obj interface{}, whatObj string, logger *zap.SugaredLogger) {
	if obj == nil {
		logger.Error(fmt.Printf("Value is nil: %s", whatObj))
		os.Exit(-1)
	}
}
