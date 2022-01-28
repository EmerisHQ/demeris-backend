package tests

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	utils "github.com/allinbits/demeris-backend/test_utils"
	"github.com/allinbits/emeris-utils/logging"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

const baseUrl = "%s://%s%s"

type testCtx struct {
	suite.Suite
	emIngress utils.EmerisIngress
	client    *http.Client
	chains    []utils.EnvChain
}

func (suite *testCtx) SetupTest() {
	logger := logging.New(logging.LoggingConfig{
		LogPath: "",
		Debug:   true,
	})

	env := os.Getenv("ENV")
	if env == "" {
		logger.Error(fmt.Printf("got nil ENV env"))
		os.Exit(-1)
	}

	emIngress, _, err := utils.LoadIngressInfo(env)
	checkNoError(err, logger)

	suite.emIngress = emIngress

	chains, err := utils.LoadChainsInfo(env)
	checkNoError(err, logger)

	suite.chains = chains

	client, err := utils.CreateNetClient(env)
	checkNoError(err, logger)

	suite.client = client
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(testCtx))
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
