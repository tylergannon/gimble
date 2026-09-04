package engine

import (
	"strings"
	"testing"

	"github.com/tylergannon/tractor/graph"
)

func TestResolveGraphModelsUsesAtomicPrecedenceAndNamedLoopRoles(t *testing.T) {
	pipeline, err := graph.Parse([]byte(`{
  "defaults":{"model":{"name":"fable","version":"5","effort":"low"}},
  "start":"worker",
  "nodes":[
    {"id":"worker","type":"agent","model":{"name":"flash"},"edges":[{"to":"fan"}]},
    {"id":"fan","type":"fan_out","model":{"name":"gpt-5.6-sol","effort":"medium"},"branches":[
      {"id":"default_branch","artifacts":["a.txt"]},
      {"id":"override_branch","artifacts":["b.txt"],"agent":{"model":{"name":"fable"}}}
    ],"branch_edges":[{"to":"join"}]},
    {"id":"join","type":"fan_in","edges":[{"to":"items"}]},
    {"id":"items","type":"loop","edges":{"loop":"worker","exit":"success"}},
    {"id":"coach","type":"supervisor","prompt":"watch","supervises":["worker"]}
  ]
}`))
	if err != nil {
		t.Fatal(err)
	}
	resolved, err := ResolveGraphModels(*pipeline, SystemModelSelection{Name: "claude-sonnet-5", Effort: "high"})
	if err != nil {
		t.Fatal(err)
	}
	byRole := map[string]ModelResolution{}
	for _, selection := range resolved {
		byRole[selection.NodeID+"/"+selection.Role] = selection
	}
	assertResolution(t, byRole["worker/agent"], "gemini-3.8-flash-medium", "medium", "gemini", "node worker agent")
	assertResolution(t, byRole["join/fan_in"], "claude-fable-5", "low", "anthropic", "pipeline defaults.model")
	assertResolution(t, byRole["coach/supervisor"], "claude-fable-5", "low", "anthropic", "pipeline defaults.model")
	assertResolution(t, byRole["default_branch/branch_agent"], "gpt-5.6-sol", "medium", "openai", "node fan fan_out template")
	// A name-only branch override is a whole replacement: neither version 5 nor
	// low effort leaks from the pipeline default.
	assertResolution(t, byRole["override_branch/branch_agent"], "claude-fable-5-1", "high", "anthropic", "node fan branch override_branch")
	assertResolution(t, byRole["items/item_judge"], "gemini-3.8-flash-medium", "medium", "gemini", "independent item_judge default")
	assertResolution(t, byRole["items/goal_evaluator"], "claude-fable-5", "low", "anthropic", "pipeline defaults.model")
}

func TestResolveGraphModelsPreflightsEveryDeclaration(t *testing.T) {
	tests := []struct {
		name string
		json string
		want string
	}{
		{name: "unused invalid default", json: `{"defaults":{"model":{"name":"unknown"}},"start":"x","nodes":[{"id":"x","type":"agent","model":{"name":"fable"},"edges":[{"to":"success"}]}]}`, want: "pipeline defaults.model"},
		{name: "hidden supervisor", json: `{"start":"x","nodes":[{"id":"x","type":"command","command":"true","edges":{"success":"success"}},{"id":"s","type":"supervisor","prompt":"watch","supervises":["x"],"model":{"name":"unknown"}}]}`, want: "node s supervisor"},
		{name: "hidden item judge", json: `{"start":"x","nodes":[{"id":"x","type":"loop","item_judge":{"model":{"name":"fable","version":"4"}},"edges":{"loop":"y","exit":"success"}},{"id":"y","type":"command","command":"true","edges":{"success":"x"}}]}`, want: "node x item_judge"},
		{name: "synthesized branch", json: `{"start":"p","nodes":[{"id":"p","type":"fan_out","branches":[{"id":"b","artifacts":["x"],"agent":{"model":{"name":"unknown"}}}],"branch_edges":[{"to":"j"}]},{"id":"j","type":"fan_in","edges":[{"to":"success"}]}]}`, want: "node p branch b"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pipeline, err := graph.Parse([]byte(test.json))
			if err != nil {
				t.Fatal(err)
			}
			_, err = ResolveGraphModels(*pipeline, SystemModelSelection{Name: "gpt-5.6-sol", Effort: "high"})
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want location %q", err, test.want)
			}
		})
	}
}

func TestSystemFallbackMatchesBetweenPreflightAndRuntime(t *testing.T) {
	pipeline, err := graph.Parse([]byte(`{"start":"worker","nodes":[{"id":"worker","type":"agent","edges":[{"to":"success"}]}]}`))
	if err != nil {
		t.Fatal(err)
	}
	node, ok := pipeline.NodeByID("worker")
	if !ok {
		t.Fatal("worker node missing")
	}
	agent := node.(*graph.AgentNode)
	tests := []struct {
		name       string
		system     SystemModelSelection
		wantModel  string
		wantEffort string
	}{
		{name: "empty config", system: SystemModelSelection{}, wantModel: "gpt-5.6-sol", wantEffort: "high"},
		{name: "Flash with default system effort", system: SystemModelSelection{Name: "flash"}, wantModel: "gemini-3.8-flash-high", wantEffort: "high"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			preflight, err := ResolveGraphModels(*pipeline, test.system)
			if err != nil {
				t.Fatal(err)
			}
			if len(preflight) != 1 {
				t.Fatalf("preflight = %+v", preflight)
			}
			runtime, err := resolveNodeModel(agent.ID, RoleAgent, agent.Model, pipeline.Defaults.Model, test.system)
			if err != nil {
				t.Fatal(err)
			}
			if runtime != preflight[0] {
				t.Fatalf("runtime = %+v, preflight = %+v", runtime, preflight[0])
			}
			if runtime.NativeModel != test.wantModel || runtime.EffectiveEffort != test.wantEffort {
				t.Fatalf("resolution = %+v, want model %q effort %q", runtime, test.wantModel, test.wantEffort)
			}
		})
	}
}

func assertResolution(t *testing.T, got ModelResolution, model, effort, provider, source string) {
	t.Helper()
	if got.NativeModel != model || got.EffectiveEffort != effort || got.Provider != provider || got.Source != source {
		t.Fatalf("resolution = %#v; want model=%q effort=%q provider=%q source=%q", got, model, effort, provider, source)
	}
}
