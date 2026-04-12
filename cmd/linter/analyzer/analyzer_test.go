package analyzer

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	// analysistest.TestData() returns path to testdata/ dir
	testdata := analysistest.TestData()

	// Run runs Analyzer on certain packages from testdata/src/
	// and checks result with comment 'want'
	analysistest.Run(t, testdata, Analyzer, "mainpkg", "otherpkg")
}
