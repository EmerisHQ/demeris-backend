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
	env          string
	emIngress    utils.EmerisIngress
	client       *http.Client
	chains       []utils.EnvChain
	clientChains []utils.EnvChain
}

func (suite *testCtx) SetupTest() {

	suite.env = os.Getenv("ENV")
	suite.Assert().NotEmpty(suite.env, "Got nil value for env:", suite.env)

	emIngress, _, err := utils.LoadIngressInfo(suite.env)
	suite.Assert().NoError(err, "err value:", err)

	suite.emIngress = emIngress

	chains, err := utils.LoadChainsInfo(suite.env)
	suite.Assert().NoError(err, "err value:", err)

	suite.chains = chains

	client, err := utils.CreateNetClient(suite.env)
	suite.Assert().NoError(err, "err value:", err)

	suite.client = client

	suite.clientChains, err = utils.LoadClientChainsInfo(suite.env)
	suite.Assert().NoError(err, "err value:", err)
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(testCtx))
}
