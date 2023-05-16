package gocapturedrefrace_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
	"gopkg.teddywing.com/gocapturedrefrace"
)

func Test(t *testing.T) {
	testdata := analysistest.TestData()

	analysistest.Run(t, testdata, gocapturedrefrace.Analyzer, ".")
}
