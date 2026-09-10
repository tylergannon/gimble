package main

import (
	"context"
	"encoding/json"
	"os"

	"github.com/tylergannon/gimble/examples/go-workflows/internal/program"
)

func main() {
	program.Main(Input{Goal: "Implement a quote CLI."}, func(ctx context.Context, in Input) error {
		rt := &program.Runtime{Agent: func(context.Context, program.Call) (any, error) { return "Canned response; no agent launched.", nil }}
		if err := ContextWalkthrough(ctx, rt, in); err != nil {
			return err
		}
		return json.NewEncoder(os.Stdout).Encode(map[string]string{"result": "canned context walkthrough"})
	})
}
