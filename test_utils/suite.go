package test_utils

import (
	"os"

	"github.com/emerishq/demeris-backend-models/cns"
	"github.com/stretchr/testify/suite"
)

type BaseTestSuite struct {
	suite.Suite
	Env       string
	EmIngress EmerisIngress
	Chains    []cns.Chain
	Client    *HttpClient
}

var allowedEnvs = []string{"dev", "staging", "prod"}

func (suite *BaseTestSuite) SetupTest() {
	suite.Env = os.Getenv("ENV")
	suite.Require().Contains(allowedEnvs, suite.Env, "ENV must be set to one of %v", allowedEnvs)

	emIngress, _, err := LoadIngressInfo(suite.Env)
	suite.Require().NoError(err)

	suite.EmIngress = emIngress

	chains, err := LoadChainsInfo(suite.Env)
	suite.Require().NoError(err)

	suite.Chains = chains

	client, err := NewHttpClient(suite.Env, suite.EmIngress.Protocol, suite.EmIngress.Host, suite.EmIngress.APIServerPath)
	suite.Require().NoError(err)

	suite.Client = client
}
