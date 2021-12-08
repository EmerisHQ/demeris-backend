package test_utils

import (
	"os"
	"runtime"
	"strings"
)

const projectName = "demeris-backend"

// Switch the working directory to the project's root
func init() {
	_, filename, _, _ := runtime.Caller(0)
	frags := strings.SplitAfter(filename, projectName)
	err := os.Chdir(frags[0])
	if err != nil {
		panic(err)
	}
}
