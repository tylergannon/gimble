//go:build integration

package workflows_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/tylergannon/tractor/checklist"
	"github.com/tylergannon/tractor/graph"
	"github.com/tylergannon/tractor/internal/workflows"
)

var (
	canonicalBinary = flag.String("tractor-binary", "", "candidate binary (default: build this checkout)")
	canonicalOutput = flag.String("proof-dir", "", "new artifact directory outside the checkout (default: temporary directory, retained)")
)

// Reuse the test executable as a native Codex launcher. This changes only
// the proof process's flags, never the operator's configuration or home.
func TestMain(m *testing.M) {
	if filepath.Base(os.Args[0]) == "codex" {
		native := os.Getenv("TRACTOR_PROOF_CODEX_EXECUTABLE")
		args := append([]string{native, "--disable", "memories"}, os.Args[1:]...)
		if err := syscall.Exec(native, args, os.Environ()); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	os.Exit(m.Run())
}

// TestCanonicalLoop runs the shipped workflow through the real CLI and native
// harnesses. No model, workflow step, application output or verdict is mocked.
func TestCanonicalLoop(t *testing.T) {
	repo, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	if status := canonicalRun(t, repo, "git", "status", "--porcelain"); len(bytes.TrimSpace(status)) != 0 {
		t.Fatalf("commit or isolate checkout changes before live proof:\n%s", status)
	}
	revision := strings.TrimSpace(string(canonicalRun(t, repo, "git", "rev-parse", "HEAD")))
	for _, name := range []string{"claude", "codex", "agy"} {
		if _, err := exec.LookPath(name); err != nil {
			t.Fatal(err)
		}
	}
	root := *canonicalOutput
	if root == "" {
		root, err = os.MkdirTemp("", "tractor-canonical-go-")
	} else {
		root, err = filepath.Abs(root)
		if err == nil && (root == repo || strings.HasPrefix(root, repo+string(os.PathSeparator))) {
			t.Fatal("proof directory must be outside the checkout")
		}
		if err == nil {
			err = os.Mkdir(root, 0o755)
		}
	}
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Proof artifacts: %s", root)
	receipt := map[string]any{"passed": false, "source_revision": revision, "started_at": time.Now().UTC()}
	canonicalJSON(t, filepath.Join(root, "result.json"), receipt)
	t.Cleanup(func() {
		receipt["passed"] = !t.Failed()
		receipt["finished_at"] = time.Now().UTC()
		canonicalJSON(t, filepath.Join(root, "result.json"), receipt)
	})
	binary := *canonicalBinary
	if binary == "" {
		binary = filepath.Join(root, "tractor")
		canonicalRun(t, repo, "go", "build", "-trimpath", "-o", binary, "./cmd/tractor")
	} else {
		binary, err = filepath.Abs(binary)
		if err != nil {
			t.Fatal(err)
		}
	}
	if err := verifyCanonicalBuild(binary, revision); err != nil {
		t.Fatal(err)
	}
	binaryHash := fmt.Sprintf("%x", sha256.Sum256(canonicalRead(t, binary)))
	receipt["binary"], receipt["binary_sha256"] = binary, binaryHash
	canonicalWrite(t, filepath.Join(root, "binary-build.txt"), canonicalRun(t, repo, "go", "version", "-m", binary))
	embedded := canonicalRun(t, repo, binary, "workflows", "show", "sprint-execute")
	want, err := workflows.Read("sprint-execute")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(embedded, want) {
		t.Fatal("candidate's embedded workflow differs from the checkout")
	}
	canonicalWrite(t, filepath.Join(root, "workflow.yaml"), embedded)

	fixture := filepath.Join(repo, "examples", "loops", "canonical")
	workspace := filepath.Join(root, "workspace")
	files := strings.Fields(string(canonicalRun(t, repo, "git", "ls-files", "--", "examples/loops/canonical")))
	for _, name := range files {
		if filepath.Base(name) == "README.md" {
			continue
		}
		rel, err := filepath.Rel(fixture, filepath.Join(repo, name))
		if err != nil {
			t.Fatal(err)
		}
		canonicalWrite(t, filepath.Join(workspace, rel), canonicalRead(t, filepath.Join(repo, name)))
	}
	canonicalRun(t, workspace, "git", "init", "-q")
	canonicalRun(t, workspace, "git", "config", "user.name", "Tractor proof")
	canonicalRun(t, workspace, "git", "config", "user.email", "tractor-proof@example.invalid")
	canonicalRun(t, workspace, "git", "add", ".")
	canonicalRun(t, workspace, "git", "-c", "core.hooksPath=/dev/null", "-c", "commit.gpgsign=false", "commit", "-qm", "Seed broken shipping CLI")
	canonicalRun(t, workspace, "git", "tag", "seed")
	// Compile the acceptance scenarios before agents run. The final check uses
	// this unchanged executable from outside the implementation workspace.
	oracle := filepath.Join(root, "shipping.test")
	canonicalRun(t, workspace, "go", "test", "-c", "-o", oracle, ".")
	canonicalShipping(t, root, workspace, oracle, "baseline", false)

	environment := canonicalEnvironment(t, root)
	args := []string{"run", "sprint-execute", "--workdir", workspace, "--logs", filepath.Join(root, "run")}
	receipt["argv"] = append([]string{binary}, args...)
	receipt["codex_overrides"] = []string{"--disable", "memories"}
	canonicalJSON(t, filepath.Join(root, "result.json"), receipt)
	t.Log("Running sprint-execute with real native agents")
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, binary, args...)
	command.Dir, command.Env = workspace, environment
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error { return syscall.Kill(-command.Process.Pid, syscall.SIGINT) }
	command.WaitDelay = 15 * time.Second
	output, err := command.CombinedOutput()
	canonicalWrite(t, filepath.Join(root, "run.log"), output)
	if err != nil {
		t.Fatalf("workflow failed: %v; inspect %s", err, root)
	}
	var manifest struct {
		Invocations []struct {
			Argv       []string `json:"argv"`
			Executable struct {
				Path    string `json:"path"`
				SHA256  string `json:"sha256"`
				Version string `json:"version"`
			} `json:"executable"`
			PipelineSource string `json:"pipeline_source"`
			GraphSHA256    string `json:"graph_sha256"`
		} `json:"invocations"`
	}
	if err := json.Unmarshal(canonicalRead(t, filepath.Join(root, "run", "manifest.json")), &manifest); err != nil {
		t.Fatal(err)
	}
	if len(manifest.Invocations) != 1 {
		t.Fatalf("run manifest invocations = %#v", manifest.Invocations)
	}
	invocation := manifest.Invocations[0]
	wantArgv := append([]string{binary}, args...)
	if !reflect.DeepEqual(invocation.Argv, wantArgv) {
		t.Fatalf("manifest argv = %q, want %q", invocation.Argv, wantArgv)
	}
	if invocation.Executable.Path != binary || invocation.Executable.SHA256 != binaryHash || invocation.Executable.Version == "" {
		t.Fatalf("manifest executable = %#v", invocation.Executable)
	}
	if invocation.PipelineSource != "builtin:sprint-execute" {
		t.Fatalf("manifest pipeline source = %q", invocation.PipelineSource)
	}
	resolved, err := graph.ParseYAML(embedded)
	if err != nil {
		t.Fatal(err)
	}
	resolvedJSON, err := json.Marshal(resolved)
	if err != nil {
		t.Fatal(err)
	}
	wantGraphHash := fmt.Sprintf("%x", sha256.Sum256(resolvedJSON))
	if invocation.GraphSHA256 != wantGraphHash {
		t.Fatalf("manifest graph hash = %q, want %q", invocation.GraphSHA256, wantGraphHash)
	}
	receipt["run_provenance"] = invocation
	events := canonicalEvents(t, filepath.Join(root, "run", "timeline.jsonl"))
	if err := verifyCanonicalTrace(events); err != nil {
		t.Fatal(err)
	}
	canonicalContract(t, fixture, workspace, files, repo)
	canonicalReviewMemory(t, filepath.Join(root, "run", "events"))
	canonicalShipping(t, root, workspace, oracle, "final", true)
	if current := fmt.Sprintf("%x", sha256.Sum256(canonicalRead(t, binary))); current != binaryHash {
		t.Fatal("candidate binary changed during proof")
	}
	if status := canonicalRun(t, repo, "git", "status", "--porcelain"); len(bytes.TrimSpace(status)) != 0 {
		t.Fatalf("checkout changed during proof:\n%s", status)
	}
	if current := strings.TrimSpace(string(canonicalRun(t, repo, "git", "rev-parse", "HEAD"))); current != revision {
		t.Fatal("checkout revision changed during proof")
	}
	receipt["exit_code"] = 0
	t.Logf("Both sprints reviewed and engine-validated; all six CLI scenarios passed. Receipt: %s", filepath.Join(root, "result.json"))
}

func canonicalShipping(t *testing.T, root, workspace, oracle, phase string, passing bool) {
	t.Helper()
	for _, mode := range []string{"standard", "expedited"} {
		name := "TestStandardShipping"
		if mode == "expedited" {
			name = "TestExpeditedShipping"
		}
		ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
		command := exec.CommandContext(ctx, oracle, "-test.run=^"+name+"$", "-test.v")
		command.Dir = workspace
		output, err := command.CombinedOutput()
		cancel()
		canonicalWrite(t, filepath.Join(root, phase+"-"+mode+".log"), output)
		if (err == nil) != passing {
			t.Fatalf("%s %s: unexpected scenario result: %v\n%s", phase, mode, err, output)
		}
		data := canonicalRead(t, filepath.Join(workspace, "evidence-"+mode+".json"))
		canonicalWrite(t, filepath.Join(root, phase+"-"+mode+".json"), data)
		var records []struct {
			Passed bool `json:"passed"`
		}
		if err := json.Unmarshal(data, &records); err != nil {
			t.Fatal(err)
		}
		if len(records) != 3 {
			t.Fatalf("%s %s: expected three real CLI invocations", phase, mode)
		}
		for index, record := range records {
			want := passing || (mode == "standard" && index != 1)
			if record.Passed != want {
				t.Fatalf("%s %s case %d: passed=%t want=%t", phase, mode, index, record.Passed, want)
			}
		}
	}
}

func canonicalContract(t *testing.T, fixture, workspace string, files []string, repo string) {
	t.Helper()
	before, err := checklist.Load(filepath.Join(fixture, "docs/sprints/ledger.md"))
	if err != nil {
		t.Fatal(err)
	}
	after, err := checklist.Load(filepath.Join(workspace, "docs/sprints/ledger.md"))
	if err != nil {
		t.Fatal(err)
	}
	if len(after.Items) != 2 {
		t.Fatal("expected two ledger items")
	}
	for index := range after.Items {
		if !after.Items[index].Done {
			t.Fatalf("item %q remains open", after.Items[index].Name)
		}
		after.Items[index].Done, after.Items[index].DonePresent = false, false
	}
	if !reflect.DeepEqual(before.Items, after.Items) || before.Body != after.Body {
		t.Fatal("acceptance ledger changed")
	}
	for _, name := range files {
		rel, err := filepath.Rel(fixture, filepath.Join(repo, name))
		if err != nil {
			t.Fatal(err)
		}
		if rel == "README.md" || rel == "cmd/quote/main.go" || rel == "docs/sprints/ledger.md" {
			continue
		}
		if !bytes.Equal(canonicalRead(t, filepath.Join(repo, name)), canonicalRead(t, filepath.Join(workspace, rel))) {
			t.Fatalf("acceptance file changed: %s", rel)
		}
	}
	if changed := strings.TrimSpace(string(canonicalRun(t, workspace, "git", "diff", "--name-only", "seed", "HEAD"))); changed != "cmd/quote/main.go" {
		t.Fatalf("agent commits changed files outside the application: %s", changed)
	}
}

func canonicalEnvironment(t *testing.T, root string) []string {
	t.Helper()
	native, err := exec.LookPath("codex")
	if err != nil {
		t.Fatal(err)
	}
	features := canonicalRun(t, root, native, "--disable", "memories", "features", "list")
	if !regexp.MustCompile(`(?m)^memories\s+\S+\s+false$`).Match(features) {
		t.Fatal("installed Codex cannot disable memories")
	}
	canonicalWrite(t, filepath.Join(root, "codex-features.txt"), features)
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	bin := filepath.Join(root, "bin")
	if err := os.Mkdir(bin, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(self, filepath.Join(bin, "codex")); err != nil {
		t.Fatal(err)
	}
	return append(os.Environ(), "PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"), "TRACTOR_PROOF_CODEX_EXECUTABLE="+native)
}

func canonicalReviewMemory(t *testing.T, root string) {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(root, "*.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range paths {
		decoder := json.NewDecoder(bytes.NewReader(canonicalRead(t, path)))
		for {
			var event struct {
				NodeID string          `json:"node_id"`
				Type   string          `json:"type"`
				Args   json.RawMessage `json:"args"`
			}
			if err := decoder.Decode(&event); err == io.EOF {
				break
			} else if err != nil {
				t.Fatal(err)
			}
			if event.NodeID == "review" && event.Type == "tool_call" && regexp.MustCompile(`\.codex/memories|MEMORY\.md|memory_summary\.md|rollout_summaries`).Match(event.Args) {
				t.Fatalf("reviewer consulted prior memory: %s", path)
			}
		}
	}
}

func canonicalEvents(t *testing.T, path string) []canonicalEvent {
	t.Helper()
	decoder := json.NewDecoder(bytes.NewReader(canonicalRead(t, path)))
	var events []canonicalEvent
	for {
		var event canonicalEvent
		if err := decoder.Decode(&event); err == io.EOF {
			return events
		} else if err != nil {
			t.Fatal(err)
		}
		events = append(events, event)
	}
}

func canonicalRun(t *testing.T, dir, name string, args ...string) []byte {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, name, args...)
	command.Dir = dir
	out, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v: %v\n%s", name, args, err, out)
	}
	return out
}

func canonicalRead(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func canonicalWrite(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func canonicalJSON(t *testing.T, path string, value any) {
	t.Helper()
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	canonicalWrite(t, path, append(data, '\n'))
}
