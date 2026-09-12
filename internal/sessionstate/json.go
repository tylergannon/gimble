// Ported from OpenCode c55ee2a8152603f04a409163bd3edf79c425fbd7. MIT,
// Copyright (c) 2025 opencode. See NOTICE.md beside this file.

package sessionstate

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// Obj is a JSON object that preserves member order, members this port does
// not name, and the difference between an absent member and one explicitly
// set to null. The upstream updater copies whole native payloads into its
// store, so a Go shape that named only the fields the reducer reads would
// silently drop the rest.
//
// A value stored in an Obj is one of: nil (JSON null), bool, string,
// json.Number, []any, or *Obj. Numbers stay json.Number so a re-encoded
// document keeps the literal the producer wrote.
type Obj struct {
	keys []string
	vals map[string]any
}

// undefinedType is the Go stand-in for JavaScript undefined: reading a member
// that is not there yields it, and assigning it removes the member. That is
// what upstream's withDefined does after building a row from event data.
type undefinedType struct{}

var undefined any = undefinedType{}

func isUndefined(value any) bool { _, ok := value.(undefinedType); return ok }

// NewObj returns an empty object.
func NewObj() *Obj { return &Obj{vals: map[string]any{}} }

// Get returns the member, or undefined when it is absent.
func (o *Obj) Get(key string) any {
	if o == nil || o.vals == nil {
		return undefined
	}
	value, ok := o.vals[key]
	if !ok {
		return undefined
	}
	return value
}

// Has reports whether the member is present, including when it is null.
func (o *Obj) Has(key string) bool {
	if o == nil || o.vals == nil {
		return false
	}
	_, ok := o.vals[key]
	return ok
}

// Set assigns the member, keeping the position of a member that already
// exists. Assigning undefined deletes it.
func (o *Obj) Set(key string, value any) {
	if isUndefined(value) {
		o.Delete(key)
		return
	}
	if o.vals == nil {
		o.vals = map[string]any{}
	}
	if _, ok := o.vals[key]; !ok {
		o.keys = append(o.keys, key)
	}
	o.vals[key] = value
}

// Delete removes the member.
func (o *Obj) Delete(key string) {
	if o == nil || o.vals == nil {
		return
	}
	if _, ok := o.vals[key]; !ok {
		return
	}
	delete(o.vals, key)
	for i, existing := range o.keys {
		if existing == key {
			o.keys = append(o.keys[:i], o.keys[i+1:]...)
			break
		}
	}
}

// Keys returns the member names in order.
func (o *Obj) Keys() []string {
	if o == nil {
		return nil
	}
	out := make([]string, len(o.keys))
	copy(out, o.keys)
	return out
}

// Clone returns a deep copy.
func (o *Obj) Clone() *Obj {
	if o == nil {
		return nil
	}
	out := &Obj{keys: make([]string, len(o.keys)), vals: make(map[string]any, len(o.vals))}
	copy(out.keys, o.keys)
	for key, value := range o.vals {
		out.vals[key] = cloneValue(value)
	}
	return out
}

func cloneValue(value any) any {
	switch typed := value.(type) {
	case *Obj:
		return typed.Clone()
	case []any:
		out := make([]any, len(typed))
		for i, item := range typed {
			out[i] = cloneValue(item)
		}
		return out
	default:
		return value
	}
}

// MarshalJSON writes the members in their recorded order.
func (o *Obj) MarshalJSON() ([]byte, error) {
	if o == nil {
		return []byte("null"), nil
	}
	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, key := range o.keys {
		if i > 0 {
			buf.WriteByte(',')
		}
		name, err := marshalValue(key)
		if err != nil {
			return nil, err
		}
		buf.Write(name)
		buf.WriteByte(':')
		value, err := marshalValue(o.vals[key])
		if err != nil {
			return nil, err
		}
		buf.Write(value)
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

// UnmarshalJSON decodes an object, preserving member order and numbers.
func (o *Obj) UnmarshalJSON(data []byte) error {
	value, err := DecodeValue(data)
	if err != nil {
		return err
	}
	decoded, ok := value.(*Obj)
	if !ok {
		return fmt.Errorf("sessionstate: expected a JSON object")
	}
	*o = *decoded
	return nil
}

func marshalValue(value any) ([]byte, error) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(value); err != nil {
		return nil, err
	}
	return bytes.TrimRight(buf.Bytes(), "\n"), nil
}

// DecodeValue decodes any JSON value into this package's document model.
func DecodeValue(data []byte) (any, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	token, err := decoder.Token()
	if err != nil {
		return nil, err
	}
	value, err := decodeValue(decoder, token)
	if err != nil {
		return nil, err
	}
	if _, err := decoder.Token(); err == nil {
		return nil, fmt.Errorf("sessionstate: trailing JSON after value")
	}
	return value, nil
}

func decodeValue(decoder *json.Decoder, token json.Token) (any, error) {
	delim, ok := token.(json.Delim)
	if !ok {
		return token, nil
	}
	switch delim {
	case '{':
		return decodeObject(decoder)
	case '[':
		return decodeArray(decoder)
	}
	return nil, fmt.Errorf("sessionstate: unexpected %v", delim)
}

func decodeObject(decoder *json.Decoder) (*Obj, error) {
	out := NewObj()
	for {
		token, err := decoder.Token()
		if err != nil {
			return nil, err
		}
		if delim, ok := token.(json.Delim); ok && delim == '}' {
			return out, nil
		}
		key, ok := token.(string)
		if !ok {
			return nil, fmt.Errorf("sessionstate: object member name must be a string")
		}
		token, err = decoder.Token()
		if err != nil {
			return nil, err
		}
		value, err := decodeValue(decoder, token)
		if err != nil {
			return nil, err
		}
		out.Set(key, value)
	}
}

func decodeArray(decoder *json.Decoder) ([]any, error) {
	out := []any{}
	for {
		token, err := decoder.Token()
		if err != nil {
			return nil, err
		}
		if delim, ok := token.(json.Delim); ok && delim == ']' {
			return out, nil
		}
		value, err := decodeValue(decoder, token)
		if err != nil {
			return nil, err
		}
		out = append(out, value)
	}
}

// str reads a member that the upstream reducer uses as a string.
func str(value any) string {
	text, _ := value.(string)
	return text
}

// objOf narrows a value to an object, yielding nil for anything else. A nil
// *Obj answers Get with undefined, which is how optional chaining reads in
// the upstream source.
func objOf(value any) *Obj {
	out, _ := value.(*Obj)
	return out
}

func arrOf(value any) []any {
	out, _ := value.([]any)
	return out
}

// truthy mirrors JavaScript truthiness for the guards the port carries over.
func truthy(value any) bool {
	switch typed := value.(type) {
	case nil:
		return false
	case undefinedType:
		return false
	case bool:
		return typed
	case string:
		return typed != ""
	case json.Number:
		number, err := typed.Float64()
		return err == nil && number != 0
	default:
		return true
	}
}

// coalesce mirrors JavaScript ??, which falls back on null and undefined only.
func coalesce(value any, fallback any) any {
	if value == nil || isUndefined(value) {
		return fallback
	}
	return value
}

// jsonKey renders values as a JSON array the way upstream compares a form's
// location against an event's, where undefined encodes as null.
func jsonKey(values ...any) string {
	items := make([]any, len(values))
	for i, value := range values {
		if isUndefined(value) {
			items[i] = nil
			continue
		}
		items[i] = value
	}
	encoded, err := marshalValue(items)
	if err != nil {
		return ""
	}
	return string(encoded)
}
