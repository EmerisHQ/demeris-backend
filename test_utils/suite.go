package test_utils

import (
	"net/http"
	"os"

	"github.com/stretchr/testify/suite"
)

type BaseTestSuite struct {
	suite.Suite
	Env       string
	EmIngress EmerisIngress
	Chains    []EnvChain
	Client    *http.Client
}

func (suite *BaseTestSuite) SetupTest() {
	suite.Env = os.Getenv("ENV")
	suite.Assert().NotEmpty(suite.Env, "Got nil value for env:", suite.Env)

	emIngress, _, err := LoadIngressInfo(suite.Env)
	suite.Assert().NoError(err, "err value:", err)

	suite.EmIngress = emIngress

	chains, err := LoadChainsInfo(suite.Env)
	suite.Assert().NoError(err, "err value:", err)

	suite.Chains = chains

	client, err := CreateNetClient(suite.Env)
	suite.Assert().NoError(err, "err value:", err)

	suite.Client = client
}
