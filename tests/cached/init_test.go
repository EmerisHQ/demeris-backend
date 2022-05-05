package tests

import (
	"testing"

	utils "github.com/emerishq/demeris-backend/test_utils"
	"github.com/stretchr/testify/suite"
)

type testCtx struct {
	utils.BaseTestSuite
}

func (suite *testCtx) SetupTest() {
	suite.BaseTestSuite.SetupTest()

	// placeholder for extending setup of this specific test suite
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(testCtx))
}
