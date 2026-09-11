//go:build jsonschema

package gimble

import (
	"encoding/json"

	"github.com/tylergannon/polytype"
)

// Stubs so the package compiles before generation; jsonschema_gen.go
// provides the real methods.
func (Review) Schema() json.RawMessage     { panic("not implemented") }
func (Review) ValidateJSON(_ []byte) error { panic("not implemented") }
func (plan) Schema() json.RawMessage       { panic("not implemented") }
func (plan) ValidateJSON(_ []byte) error   { panic("not implemented") }

var (
	_ = polytype.Declare(Review.Schema)
	_ = polytype.Declare(plan.Schema)
)
