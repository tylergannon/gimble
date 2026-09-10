// Command contextcheck preserves an unadopted intraprocedural experiment.
// It must not be used as a workflow ownership validator.
package main

import (
	"fmt"
	"os"

	"github.com/tylergannon/gimble/examples/go-workflows/internal/contextcheck"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	fmt.Fprintln(os.Stderr, "Unadopted experiment: contextcheck does not traverse the call graph. A successful exit does not validate workflow ownership.")
	singlechecker.Main(contextcheck.Analyzer)
}
