package test_utils

import (
	"os"
	"runtime"
	"strings"
)

const projectName = "demeris-backend"

// Switch the working directory to the project's root
func init() {
	overridePwd := os.Getenv("TESTS_WORKDIR")
	if len(overridePwd) == 0 {
		_, filename, _, _ := runtime.Caller(0)
		frags := strings.SplitAfter(filename, projectName)
		overridePwd = frags[0]
	}

	err := os.Chdir(overridePwd)
	if err != nil {
		panic(err)
	}
}
