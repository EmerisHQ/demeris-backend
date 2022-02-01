package tests

import (
	"net/http"
	"os"
	"testing"

	utils "github.com/allinbits/demeris-backend/test_utils"
	"github.com/stretchr/testify/suite"
)

const baseUrl = "%s://%s%s"

type testCtx struct {
	suite.Suite
	emIngress utils.EmerisIngress
	client    *http.Client
	chains    []utils.EnvChain
}

func (suite *testCtx) SetupTest() {

	env := os.Getenv("ENV")
	suite.Assert().NotEmpty(env, "Got nil value for env:", env)

	emIngress, _, err := utils.LoadIngressInfo(env)
	suite.Assert().NoError(err, "err value:", err)

	suite.emIngress = emIngress

	chains, err := utils.LoadChainsInfo(env)
	suite.Assert().NoError(err, "err value:", err)

	suite.chains = chains

	client, err := utils.CreateNetClient(env)
	suite.Assert().NoError(err, "err value:", err)

	suite.client = client
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(testCtx))
}
