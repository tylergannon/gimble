package program

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
)

// Main keeps JSON/CLI plumbing out of the workflow. T is that workflow's own
// argument shape; schema generation and a builtin catalog are separate work.
func Main[T any](example T, run func(context.Context, T) error) {
	inputPath := flag.String("input", "", "JSON argument file (- reads stdin); defaults to the example")
	showExample := flag.Bool("example", false, "print an example JSON argument and exit")
	flag.Parse()
	if *showExample {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		exitOnError(encoder.Encode(example))
		return
	}
	if flag.NArg() != 0 {
		exitOnError(fmt.Errorf("unexpected arguments: %v", flag.Args()))
	}
	input := example
	if *inputPath != "" {
		var reader io.Reader = os.Stdin
		if *inputPath != "-" {
			file, err := os.Open(*inputPath)
			exitOnError(err)
			defer func() { _ = file.Close() }()
			reader = file
		}
		var decoded T // An explicit input replaces the example; it never overlays it.
		decoder := json.NewDecoder(reader)
		decoder.DisallowUnknownFields()
		exitOnError(decoder.Decode(&decoded))
		if err := decoder.Decode(new(any)); !errors.Is(err, io.EOF) {
			exitOnError(fmt.Errorf("expected one JSON argument, then end of input"))
		}
		input = decoded
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	exitOnError(run(ctx, input))
}

func exitOnError(err error) {
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
