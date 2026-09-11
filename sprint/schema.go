//go:build jsonschema

package sprint

import (
	"encoding/json"

	"github.com/tylergannon/polytype"
)

// Stubs so the package compiles before generation; jsonschema_gen.go
// provides the real methods.
func (Input) Schema() json.RawMessage       { panic("not implemented") }
func (Input) ValidateJSON(_ []byte) error   { panic("not implemented") }
func (verdict) Schema() json.RawMessage     { panic("not implemented") }
func (verdict) ValidateJSON(_ []byte) error { panic("not implemented") }

var (
	_ = polytype.Declare(Input.Schema)
	_ = polytype.Declare(verdict.Schema)
)
