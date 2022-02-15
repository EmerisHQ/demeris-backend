package tests

import (
	"testing"

	utils "github.com/allinbits/demeris-backend/test_utils"
	"github.com/stretchr/testify/suite"
)

const baseUrl = "%s://%s%s"

type testCtx struct {
	utils.BaseTestSuite

	clientChains []utils.EnvChain
}

func (suite *testCtx) SetupTest() {
	suite.BaseTestSuite.SetupTest()

	var err error
	suite.clientChains, err = utils.LoadClientChainsInfo(suite.Env)
	suite.Require().NoError(err, "err value:", err)
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(testCtx))
}
