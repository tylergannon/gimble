package contextcheck

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestOwnership(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "github.com/tylergannon/gimble/examples/go-workflows/checkcases")
}
