package main

import (
	"context"
	"encoding/json"
	"os"

	"github.com/tylergannon/gimble/examples/go-workflows/internal/program"
)

func main() {
	program.Main(exampleInput(), func(ctx context.Context, in Input) error {
		rt := &program.Runtime{Agent: func(context.Context, program.Call) (any, error) { return "Canned response; no agent launched.", nil }}
		items, err := NestedScopes(ctx, rt, in)
		if err != nil {
			return err
		}
		return json.NewEncoder(os.Stdout).Encode(map[string]any{"result": "canned nested scopes", "items": items})
	})
}
