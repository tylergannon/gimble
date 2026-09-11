//go:build jsonschema

package gimble

import (
	"encoding/json"

	"github.com/tylergannon/polytype"
)

// Stubs so the package compiles before generation; jsonschema_gen.go
// provides the real methods.
func (review) Schema() json.RawMessage              { panic("not implemented") }
func (review) ValidateJSON(_ []byte) error          { panic("not implemented") }
func (plan) Schema() json.RawMessage                { panic("not implemented") }
func (plan) ValidateJSON(_ []byte) error            { panic("not implemented") }
func (lifecycleRecord) Schema() json.RawMessage     { panic("not implemented") }
func (lifecycleRecord) ValidateJSON(_ []byte) error { panic("not implemented") }
func (agentRecord) Schema() json.RawMessage         { panic("not implemented") }
func (agentRecord) ValidateJSON(_ []byte) error     { panic("not implemented") }

var (
	_ = polytype.Declare(review.Schema)
	_ = polytype.Declare(plan.Schema)
	_ = polytype.Declare(lifecycleRecord.Schema)
	_ = polytype.Declare(agentRecord.Schema)
	_ = polytype.SealedUnion[lifecycleEvent]("kind", polytype.Snake)
	_ = polytype.SealedUnion[AgentEvent]("kind", polytype.Snake)
)
