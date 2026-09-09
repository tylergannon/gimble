package shipping_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

type quote struct {
	Subtotal int    `json:"subtotal_cents"`
	Shipping int    `json:"shipping_cents"`
	Total    int    `json:"total_cents"`
	Mode     string `json:"mode"`
}

type invocation struct {
	Args     []string `json:"args"`
	ExitCode int      `json:"exit_code"`
	Stdout   string   `json:"stdout"`
	Stderr   string   `json:"stderr"`
	Expected quote    `json:"expected"`
	Passed   bool     `json:"passed"`
}

func TestStandardShipping(t *testing.T) {
	checkShipping(t, "standard")
}

func TestExpeditedShipping(t *testing.T) {
	checkShipping(t, "expedited")
}

func checkShipping(t *testing.T, mode string) {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "quote")
	ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
	defer cancel()
	if out, err := exec.CommandContext(ctx, "go", "build", "-o", binary, "./cmd/quote").CombinedOutput(); err != nil {
		t.Fatalf("build quote: %v\n%s", err, out)
	}
	var records []invocation
	for _, cents := range []int{4999, 5000, 5001} {
		amount := fmt.Sprintf("%d.%02d", cents/100, cents%100)
		t.Run(amount, func(t *testing.T) {
			shipping := 800
			if mode == "expedited" {
				shipping = 1200
			} else if cents >= 5000 {
				shipping = 0
			}
			record := invocation{Args: []string{amount, mode}, Expected: quote{cents, shipping, cents + shipping, mode}}
			ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
			defer cancel()
			command := exec.CommandContext(ctx, binary, record.Args...)
			var stdout, stderr bytes.Buffer
			command.Stdout, command.Stderr = &stdout, &stderr
			err := command.Run()
			record.Stdout, record.Stderr = stdout.String(), stderr.String()
			if err != nil {
				record.ExitCode = -1
				if command.ProcessState != nil {
					record.ExitCode = command.ProcessState.ExitCode()
				}
			}
			var actual quote
			decoder := json.NewDecoder(&stdout)
			decoder.DisallowUnknownFields()
			decodeErr := decoder.Decode(&actual)
			var extra any
			record.Passed = err == nil && decodeErr == nil && actual == record.Expected && decoder.Decode(&extra) == io.EOF
			records = append(records, record)
			if !record.Passed {
				t.Errorf("CLI %v: exit=%d stdout=%q stderr=%q; want %+v", record.Args, record.ExitCode, record.Stdout, record.Stderr, record.Expected)
			}
		})
	}
	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("evidence-"+mode+".json", append(data, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
}
