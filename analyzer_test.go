package analyzer_test

import (
	"testing"

	"github.com/nitinmewar/alias-analyzer/analyzer"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestSliceAliasAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, analyzer.Analyzer, "slicetest")
}
