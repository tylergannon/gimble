package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNextQuestionPathUsesHighestExistingNumber(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"0002.html", "0007.md", "0009.answer.md", "notes.md"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("test"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	path, err := nextQuestionPath(dir, ".md")
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(dir, "0008.md"); path != want {
		t.Fatalf("next question = %q, want %q", path, want)
	}
}

func TestPrepareQuestionPreservesSupportedExtensionAndRejectsOthers(t *testing.T) {
	dir := t.TempDir()
	html := filepath.Join(t.TempDir(), "question.html")
	if err := os.WriteFile(html, []byte("<p>Question</p>"), 0o644); err != nil {
		t.Fatal(err)
	}
	moved, resumed, err := prepareQuestion(html, dir)
	if err != nil {
		t.Fatal(err)
	}
	if resumed || moved != filepath.Join(dir, "0001.html") {
		t.Fatalf("prepareQuestion = %q, %v", moved, resumed)
	}
	if _, err := os.Stat(html); !os.IsNotExist(err) {
		t.Fatalf("source still exists or stat failed unexpectedly: %v", err)
	}

	text := filepath.Join(t.TempDir(), "question.txt")
	if err := os.WriteFile(text, []byte("Question"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := prepareQuestion(text, dir); err == nil || !strings.Contains(err.Error(), ".md or .html") {
		t.Fatalf("unsupported extension error = %v", err)
	}
	if _, err := os.Stat(text); err != nil {
		t.Fatalf("rejected source was moved: %v", err)
	}
}

func TestAskResumesNumberedQuestionWithoutRenumbering(t *testing.T) {
	dir := t.TempDir()
	question := filepath.Join(dir, "0001.md")
	if err := os.WriteFile(question, []byte("Question"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(answerPathForQuestion(question), []byte("Answer"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GIMBLE_INTERVIEW_DIR", dir)
	t.Setenv("GIMBLE_RUN_DIR", "")

	stdout, stderr, err := executeCommand("ask", question)
	if err != nil {
		t.Fatal(err)
	}
	if stdout != "Answer" {
		t.Fatalf("stdout = %q", stdout)
	}
	if stderr != "" {
		t.Fatalf("resume stderr = %q", stderr)
	}
	if _, err := os.Stat(filepath.Join(dir, "0002.md")); !os.IsNotExist(err) {
		t.Fatalf("resume created another question: %v", err)
	}
}

func TestAskAppendsQuestionAskedTimelineEvent(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "interview")
	runDir := filepath.Join(root, "run")
	if err := os.Mkdir(runDir, 0o755); err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(root, "question.md")
	if err := os.WriteFile(source, []byte("Question"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GIMBLE_INTERVIEW_DIR", dir)
	t.Setenv("GIMBLE_RUN_DIR", runDir)

	command := newRootCommand()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.SetOut(&stdout)
	command.SetErr(&stderr)
	command.SetArgs([]string{"ask", source})
	done := make(chan error, 1)
	go func() { done <- command.Execute() }()

	question := filepath.Join(dir, "0001.md")
	waitForInterviewFile(t, question)
	answerCommand := newRootCommand()
	answerCommand.SetOut(&bytes.Buffer{})
	answerCommand.SetErr(&bytes.Buffer{})
	answerCommand.SetArgs([]string{"answer", question, "Reviewer answer"})
	if err := answerCommand.Execute(); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("ask did not return after answer was written")
	}
	if stdout.String() != "Reviewer answer" {
		t.Fatalf("stdout = %q", stdout.String())
	}

	raw, err := os.ReadFile(filepath.Join(runDir, "timeline.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	var event map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(raw), &event); err != nil {
		t.Fatal(err)
	}
	if event["type"] != "QuestionAsked" || event["question"] != question || event["ts"] == "" {
		t.Fatalf("timeline event = %#v", event)
	}
}

func TestAnswerRefusesToOverwrite(t *testing.T) {
	dir := t.TempDir()
	question := filepath.Join(dir, "0001.md")
	if err := os.WriteFile(question, []byte("Question"), 0o644); err != nil {
		t.Fatal(err)
	}
	stdout, _, err := executeCommand("answer", question, "first")
	if err != nil {
		t.Fatal(err)
	}
	answerPath := answerPathForQuestion(question)
	if stdout != answerPath+"\n" {
		t.Fatalf("stdout = %q", stdout)
	}
	if _, _, err := executeCommand("answer", question, "second"); err == nil || !strings.Contains(err.Error(), "refusing to overwrite") {
		t.Fatalf("second answer error = %v", err)
	}
	raw, err := os.ReadFile(answerPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != "first" {
		t.Fatalf("answer changed to %q", raw)
	}
}

func TestAskRequiresInterviewDirectory(t *testing.T) {
	t.Setenv("GIMBLE_INTERVIEW_DIR", "")
	_, _, err := executeCommand("ask", "question.md")
	if err == nil || !strings.Contains(err.Error(), "--into") || !strings.Contains(err.Error(), "GIMBLE_INTERVIEW_DIR") {
		t.Fatalf("error = %v", err)
	}
}

func waitForInterviewFile(t *testing.T, path string) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(path); err == nil {
			return
		} else if !os.IsNotExist(err) {
			t.Fatal(err)
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("file %q was not created", path)
}
