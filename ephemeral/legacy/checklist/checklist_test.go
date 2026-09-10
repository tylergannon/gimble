package checklist

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const example = `---
items:
  - name: Build the login screen
    check: The login screen validates and submits on valid input
    command: npx playwright test tests/login.spec.ts
    infer:
      files: ephemeral/captures/login/*.png
      prompt: Judge whether every state of the login screen looks usable
    doc: ephemeral/projects/mvp/sprints/SPRINT-0002.md
  - name: Reject a bad password
    check: A wrong password shows an inline error and keeps the form
    command: npx playwright test tests/login-error.spec.ts
    done: true
---

# Sprint 2: login

Body prose. Never parsed.
`

func TestParseExample(t *testing.T) {
	list, err := Parse("sprint.md", []byte(example))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if list.Path != "sprint.md" {
		t.Errorf("Path = %q", list.Path)
	}
	if len(list.Items) != 2 {
		t.Fatalf("got %d items, want 2", len(list.Items))
	}

	first := list.Items[0]
	if first.Name != "Build the login screen" {
		t.Errorf("first.Name = %q", first.Name)
	}
	if first.Check != "The login screen validates and submits on valid input" {
		t.Errorf("first.Check = %q", first.Check)
	}
	if first.Command != "npx playwright test tests/login.spec.ts" {
		t.Errorf("first.Command = %q", first.Command)
	}
	if first.Infer == nil {
		t.Fatal("first.Infer is nil")
	}
	if got, want := first.Infer.Files, []string{"ephemeral/captures/login/*.png"}; !equal(got, want) {
		t.Errorf("first.Infer.Files = %v, want %v", got, want)
	}
	if first.Infer.Prompt != "Judge whether every state of the login screen looks usable" {
		t.Errorf("first.Infer.Prompt = %q", first.Infer.Prompt)
	}
	if first.Doc != "ephemeral/projects/mvp/sprints/SPRINT-0002.md" {
		t.Errorf("first.Doc = %q", first.Doc)
	}
	if first.Done {
		t.Error("first.Done should be false")
	}
	if first.DonePresent {
		t.Error("first.DonePresent should be false")
	}

	second := list.Items[1]
	if second.Name != "Reject a bad password" {
		t.Errorf("second.Name = %q", second.Name)
	}
	if !second.Done {
		t.Error("second.Done should be true")
	}
	if !second.DonePresent {
		t.Error("second.DonePresent should be true")
	}
	if second.Infer != nil {
		t.Error("second.Infer should be nil")
	}

	wantBody := "\n\n# Sprint 2: login\n\nBody prose. Never parsed.\n"
	if list.Body != wantBody {
		t.Errorf("Body = %q, want %q", list.Body, wantBody)
	}
}

func TestParseInferFilesForms(t *testing.T) {
	single := "---\nitems:\n  - name: a\n    check: c\n    infer:\n      files: one/*.png\n      prompt: p\n---\n"
	list, err := Parse("x.md", []byte(single))
	if err != nil {
		t.Fatalf("single: %v", err)
	}
	if got := list.Items[0].Infer.Files; !equal(got, []string{"one/*.png"}) {
		t.Errorf("single files = %v", got)
	}

	multi := "---\nitems:\n  - name: a\n    check: c\n    infer:\n      files:\n        - one/*.png\n        - two/*.png\n      prompt: p\n---\n"
	list, err = Parse("x.md", []byte(multi))
	if err != nil {
		t.Fatalf("multi: %v", err)
	}
	if got := list.Items[0].Infer.Files; !equal(got, []string{"one/*.png", "two/*.png"}) {
		t.Errorf("multi files = %v", got)
	}
}

func TestParseKeepsExtraTopLevelKeys(t *testing.T) {
	src := "---\ntitle: Sprint 2\nowner: tyler\nitems:\n  - name: a\n    check: c\n---\nbody\n"
	list, err := Parse("x.md", []byte(src))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(list.Items) != 1 || list.Body != "\nbody\n" {
		t.Errorf("unexpected parse result: %+v", list)
	}
}

func TestParseEmptyItems(t *testing.T) {
	list, err := Parse("x.md", []byte("---\nitems: []\n---\n"))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(list.Items) != 0 {
		t.Errorf("got %d items", len(list.Items))
	}
	if _, _, ok := list.Open(); ok {
		t.Error("Open on empty list should be ok=false")
	}
}

func TestOpen(t *testing.T) {
	list := &Checklist{Items: []Item{
		{Name: "a", Done: true},
		{Name: "b"},
		{Name: "c"},
	}}
	item, index, ok := list.Open()
	if !ok || item.Name != "b" || index != 1 {
		t.Errorf("Open = (%q, %d, %v), want (b, 1, true)", item.Name, index, ok)
	}

	all := &Checklist{Items: []Item{{Name: "a", Done: true}, {Name: "b", Done: true}}}
	if _, _, ok := all.Open(); ok {
		t.Error("Open should be ok=false when everything is done")
	}
}

func TestFind(t *testing.T) {
	list := &Checklist{Items: []Item{{Name: "a"}, {Name: "b"}}}
	item, index, ok := list.Find("b")
	if !ok || index != 1 || item.Name != "b" {
		t.Errorf("Find(b) = (%q, %d, %v)", item.Name, index, ok)
	}
	if _, _, ok := list.Find("zzz"); ok {
		t.Error("Find of unknown name should be ok=false")
	}
}

const commented = `---
title: Sprint 2
items:
  # The screen itself comes first.
  - name: Build the login screen
    check: The login screen validates and submits on valid input
    doc: ephemeral/projects/mvp/sprints/SPRINT-0002.md
    command: npx playwright test tests/login.spec.ts
  - name: Reject a bad password
    check: A wrong password shows an inline error and keeps the form
    done: false
---

# Sprint 2: login

Body prose with trailing spaces.
	And a tab-indented line.
No trailing newline`

func TestMarkDone(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sprint.md")
	if err := os.WriteFile(path, []byte(commented), 0o644); err != nil {
		t.Fatal(err)
	}
	before, err := Load(path)
	if err != nil {
		t.Fatalf("Load before: %v", err)
	}

	if err := MarkDone(path, "Build the login screen"); err != nil {
		t.Fatalf("MarkDone: %v", err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(raw)
	if !strings.HasSuffix(text, before.Body) {
		t.Errorf("body not preserved byte for byte:\n%s", text)
	}
	if !strings.Contains(text, "# The screen itself comes first.") {
		t.Errorf("YAML comment lost:\n%s", text)
	}
	if !strings.HasPrefix(text, "---\ntitle: Sprint 2\nitems:\n") {
		t.Errorf("top-level key order not preserved:\n%s", text)
	}
	docIdx := strings.Index(text, "doc:")
	cmdIdx := strings.Index(text, "command:")
	if docIdx < 0 || cmdIdx < 0 || docIdx > cmdIdx {
		t.Errorf("item key order not preserved (doc before command):\n%s", text)
	}

	after, err := Load(path)
	if err != nil {
		t.Fatalf("Load after: %v", err)
	}
	if after.Body != before.Body {
		t.Errorf("Body changed:\nbefore %q\nafter  %q", before.Body, after.Body)
	}
	if !after.Items[0].Done {
		t.Error("first item not Done after MarkDone")
	}
	if after.Items[1].Done {
		t.Error("second item should still be open")
	}
	if after.Items[0].Command != before.Items[0].Command || after.Items[0].Doc != before.Items[0].Doc {
		t.Error("first item fields changed on rewrite")
	}
	item, index, ok := after.Open()
	if !ok || index != 1 || item.Name != "Reject a bad password" {
		t.Errorf("Open after mark = (%q, %d, %v)", item.Name, index, ok)
	}

	// Marking the existing done: false key flips it in place.
	if err := MarkDone(path, "Reject a bad password"); err != nil {
		t.Fatalf("MarkDone second: %v", err)
	}
	after, err = Load(path)
	if err != nil {
		t.Fatalf("Load after second: %v", err)
	}
	if _, _, ok := after.Open(); ok {
		t.Error("all items should be done")
	}
	raw, _ = os.ReadFile(path)
	if strings.Count(string(raw), "done:") != 2 {
		t.Errorf("expected exactly two done keys:\n%s", raw)
	}

	// Marking an already-done item is a no-op that still succeeds.
	if err := MarkDone(path, "Reject a bad password"); err != nil {
		t.Fatalf("MarkDone already done: %v", err)
	}
	again, _ := os.ReadFile(path)
	if string(again) != string(raw) {
		t.Errorf("no-op mark changed the file:\n%s", again)
	}

	if err := MarkDone(path, "Nope"); err == nil {
		t.Error("MarkDone of unknown name should fail")
	} else if !strings.Contains(err.Error(), "Nope") {
		t.Errorf("error should name the item: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o644 {
		t.Errorf("mode = %o, want 644", info.Mode().Perm())
	}
	entries, _ := os.ReadDir(filepath.Dir(path))
	if len(entries) != 1 {
		t.Errorf("temp files left behind: %v", entries)
	}
}

func TestUnmarkDone(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sprint.md")
	if err := os.WriteFile(path, []byte(commented), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := MarkDone(path, "Build the login screen"); err != nil {
		t.Fatalf("MarkDone: %v", err)
	}
	before, err := Load(path)
	if err != nil {
		t.Fatalf("Load before: %v", err)
	}
	if err := UnmarkDone(path, "Build the login screen"); err != nil {
		t.Fatalf("UnmarkDone: %v", err)
	}
	after, err := Load(path)
	if err != nil {
		t.Fatalf("Load after: %v", err)
	}
	if after.Items[0].Done || after.Body != before.Body {
		t.Fatalf("after unmark = %+v", after)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), "done: false") {
		t.Fatalf("unmark did not write done: false:\n%s", raw)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("mode = %o, want 600", info.Mode().Perm())
	}
	if err := UnmarkDone(path, "Nope"); err == nil || !strings.Contains(err.Error(), "Nope") {
		t.Fatalf("UnmarkDone unknown item error = %v", err)
	}
}

func TestMarkDoneRoundTripsCRLFBody(t *testing.T) {
	src := "---\r\nitems:\r\n  - name: a\r\n    check: c\r\n---\r\nbody\r\n"
	path := filepath.Join(t.TempDir(), "crlf.md")
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	before, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if before.Body != "\r\nbody\r\n" {
		t.Errorf("Body = %q", before.Body)
	}
	if err := MarkDone(path, "a"); err != nil {
		t.Fatalf("MarkDone: %v", err)
	}
	after, err := Load(path)
	if err != nil {
		t.Fatalf("Load after: %v", err)
	}
	if after.Body != before.Body || !after.Items[0].Done {
		t.Errorf("after = %+v", after)
	}
}

func TestMarkDoneThroughSymlinkMarksTarget(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "real.md")
	if err := os.WriteFile(target, []byte(example), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "link.md")
	if err := os.Symlink("real.md", link); err != nil {
		t.Fatal(err)
	}

	if err := MarkDone(link, "Build the login screen"); err != nil {
		t.Fatalf("MarkDone: %v", err)
	}

	info, err := os.Lstat(link)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatal("link was replaced by a regular file")
	}
	real, err := Load(target)
	if err != nil {
		t.Fatalf("Load target: %v", err)
	}
	if !real.Items[0].Done {
		t.Error("target was not marked done")
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 2 {
		t.Errorf("unexpected directory contents: %v", entries)
	}
}

func TestMarkDonePreservesFileMode(t *testing.T) {
	path := filepath.Join(t.TempDir(), "private.md")
	if err := os.WriteFile(path, []byte(example), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := MarkDone(path, "Build the login screen"); err != nil {
		t.Fatalf("MarkDone: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("mode = %o, want 600", info.Mode().Perm())
	}
}

func TestParseErrors(t *testing.T) {
	cases := map[string]struct {
		src  string
		want string
	}{
		"missing frontmatter": {
			src:  "# Just markdown\n",
			want: "missing frontmatter",
		},
		"unclosed frontmatter": {
			src:  "---\nitems: []\n",
			want: "no closing ---",
		},
		"no items": {
			src:  "---\ntitle: x\n---\n",
			want: "no items key",
		},
		"items not a list": {
			src:  "---\nitems: nope\n---\n",
			want: "items must be a list",
		},
		"duplicate names": {
			src:  "---\nitems:\n  - name: a\n    check: c\n  - name: a\n    check: d\n---\n",
			want: "duplicate name \"a\"",
		},
		"missing check": {
			src:  "---\nitems:\n  - name: a\n---\n",
			want: "check is required",
		},
		"blank name": {
			src:  "---\nitems:\n  - name: '  '\n    check: c\n---\n",
			want: "name is required",
		},
		"unknown item key": {
			src:  "---\nitems:\n  - name: a\n    check: c\n    comand: ls\n---\n",
			want: "unknown key \"comand\"",
		},
		"infer without prompt": {
			src:  "---\nitems:\n  - name: a\n    check: c\n    infer:\n      files: x/*.png\n---\n",
			want: "infer.prompt is required",
		},
		"infer without files": {
			src:  "---\nitems:\n  - name: a\n    check: c\n    infer:\n      prompt: p\n---\n",
			want: "infer.files must list at least one path",
		},
		"infer unknown key": {
			src:  "---\nitems:\n  - name: a\n    check: c\n    infer:\n      files: x\n      prompt: p\n      model: fast\n---\n",
			want: "unknown key \"model\"",
		},
		"done not a bool": {
			src:  "---\nitems:\n  - name: a\n    check: c\n    done: yes please\n---\n",
			want: "done must be true or false",
		},
		"item not a mapping": {
			src:  "---\nitems:\n  - just a string\n---\n",
			want: "must be a mapping",
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := Parse("bad.md", []byte(tc.src))
			if err == nil {
				t.Fatalf("expected error containing %q", tc.want)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error %q does not contain %q", err, tc.want)
			}
			if !strings.Contains(err.Error(), "bad.md") {
				t.Errorf("error %q does not name the path", err)
			}
		})
	}
}

func TestRender(t *testing.T) {
	full := Item{
		Name:      "Build the login screen",
		Check:     "The login screen validates and submits on valid input",
		Command:   "npx playwright test tests/login.spec.ts",
		Infer:     &Infer{Files: []string{"a/*.png", "b/*.png"}, Prompt: "Judge it"},
		Doc:       "docs/sprint.md",
		Checklist: "docs/sub.md",
		Done:      true,
	}
	got := full.Render()
	for _, want := range []string{
		"name: Build the login screen",
		"check: The login screen validates and submits on valid input",
		"command: npx playwright test tests/login.spec.ts",
		"infer:",
		"prompt: Judge it",
		"files:",
		"- a/*.png",
		"- b/*.png",
		"doc: docs/sprint.md",
		"checklist: docs/sub.md",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("Render missing %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "done") {
		t.Errorf("Render printed done:\n%s", got)
	}
	if strings.HasSuffix(got, "\n") {
		t.Errorf("Render should not end with a newline:\n%q", got)
	}

	minimal := Item{Name: "a", Check: "c", Done: true}.Render()
	if minimal != "name: a\ncheck: c" {
		t.Errorf("minimal Render = %q", minimal)
	}
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
