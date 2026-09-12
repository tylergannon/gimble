// Ported from OpenCode c55ee2a8152603f04a409163bd3edf79c425fbd7. MIT.
package sessionstate

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"
)

type oracleRow struct {
	Prefix    int             `json:"prefix"`
	Immediate json.RawMessage `json:"immediate"`
}

func loadOracle(t *testing.T, name string) []oracleRow {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata/oracle", name+".jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	var rows []oracleRow
	for _, line := range splitLines(string(data)) {
		var row oracleRow
		if err := json.Unmarshal([]byte(line), &row); err != nil {
			t.Fatalf("oracle %s: %v", name, err)
		}
		rows = append(rows, row)
	}
	return rows
}

func requireState(t *testing.T, label string, got ProjectionState, want json.RawMessage) {
	t.Helper()
	encoded, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(canonical(t, string(encoded)), canonical(t, string(want))) {
		t.Fatalf("%s diverges from upstream\n go: %s\nora: %s", label, encoded, want)
	}
}

func oracleFixtures() []string {
	return []string{"native-backend", "stream", "overlapping-reasoning"}
}

func newFromFixture(t *testing.T, fixture *Obj) *Projection {
	t.Helper()
	return New(decodeState(t, seedState(fixture)))
}

func TestEveryPrefixMatchesUpstreamOracle(t *testing.T) {
	for _, name := range oracleFixtures() {
		t.Run(name, func(t *testing.T) {
			fixture := loadFixture(t, name+".json")
			rows := loadOracle(t, name)
			events := arrOf(fixture.Get("events"))
			if len(rows) != len(events)+1 {
				t.Fatalf("oracle has %d prefixes for %d events", len(rows), len(events))
			}
			projection := New(decodeState(t, seedState(fixture)))
			requireState(t, name+" prefix 0", projection.Snapshot().State, rows[0].Immediate)
			for cut, entry := range events {
				projection.Apply(objOf(entry))
				requireState(t, name+" prefix "+strconv.Itoa(cut+1), projection.Snapshot().State, rows[cut+1].Immediate)
			}
		})
	}
}
