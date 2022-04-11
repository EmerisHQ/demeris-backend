package tests

import (
	"testing"

	chainClient "github.com/allinbits/demeris-backend/chain_client"
	utils "github.com/allinbits/demeris-backend/test_utils"
	"github.com/stretchr/testify/suite"
)

type testCtx struct {
	utils.BaseTestSuite

	clientChains []chainClient.ChainClient
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
