package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const answerPollInterval = 250 * time.Millisecond

var questionNamePattern = regexp.MustCompile(`^([0-9]{4})\.(md|html)$`)

func newAskCommand() *cobra.Command {
	var interviewDir string
	command := &cobra.Command{
		Use:   "ask <path>",
		Short: "Ask a question and wait for its answer",
		Long: "Move a Markdown or HTML question into the interview directory and wait for its answer.\n" +
			"The answer file is checked every 250ms until it exists and is non-empty.",
		Args: cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			dir, err := resolveInterviewDir(interviewDir, command.Flags().Changed("into"))
			if err != nil {
				return err
			}
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return fmt.Errorf("create interview directory %q: %w", dir, err)
			}

			questionPath, resumed, err := prepareQuestion(args[0], dir)
			if err != nil {
				return err
			}
			if !resumed {
				displayPath := pathForDisplay(questionPath)
				if _, err := fmt.Fprintf(command.ErrOrStderr(), "question: %s\n", displayPath); err != nil {
					return fmt.Errorf("print question path: %w", err)
				}
				if runDir := os.Getenv("TRACTOR_RUN_DIR"); runDir != "" {
					if err := appendQuestionAsked(runDir, displayPath); err != nil {
						return err
					}
				} else if _, err := fmt.Fprintln(command.ErrOrStderr(), "warning: TRACTOR_RUN_DIR is unset; question was not recorded in a run timeline"); err != nil {
					return fmt.Errorf("print timeline warning: %w", err)
				}
			}

			answerPath := answerPathForQuestion(questionPath)
			answer, err := waitForAnswer(command, answerPath)
			if err != nil {
				return err
			}
			if _, err := command.OutOrStdout().Write(answer); err != nil {
				return fmt.Errorf("print answer: %w", err)
			}
			return nil
		},
	}
	command.Flags().StringVar(&interviewDir, "into", "", "interview directory (default TRACTOR_INTERVIEW_DIR)")
	return command
}

func newAnswerCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "answer <question-path> [text]",
		Short: "Answer a waiting question",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(command *cobra.Command, args []string) error {
			questionPath, err := validQuestionPath(args[0])
			if err != nil {
				return err
			}
			var answer []byte
			if len(args) == 2 {
				answer = []byte(args[1])
			} else {
				answer, err = readAnswerInput(command.InOrStdin())
				if err != nil {
					return err
				}
			}
			if len(answer) == 0 {
				return fmt.Errorf("answer text is required as an argument or on stdin")
			}
			answerPath := answerPathForQuestion(questionPath)
			if err := writeExclusive(answerPath, answer); err != nil {
				return err
			}
			_, err = fmt.Fprintln(command.OutOrStdout(), answerPath)
			return err
		},
	}
}

func resolveInterviewDir(flagValue string, flagSet bool) (string, error) {
	if flagSet {
		if flagValue == "" {
			return "", fmt.Errorf("--into requires a directory")
		}
		return flagValue, nil
	}
	if dir := os.Getenv("TRACTOR_INTERVIEW_DIR"); dir != "" {
		return dir, nil
	}
	return "", fmt.Errorf("interview directory is required: use --into or set TRACTOR_INTERVIEW_DIR")
}

func prepareQuestion(sourcePath, interviewDir string) (string, bool, error) {
	sourceAbs, err := filepath.Abs(sourcePath)
	if err != nil {
		return "", false, fmt.Errorf("resolve question path %q: %w", sourcePath, err)
	}
	dirAbs, err := filepath.Abs(interviewDir)
	if err != nil {
		return "", false, fmt.Errorf("resolve interview directory %q: %w", interviewDir, err)
	}
	if filepath.Dir(sourceAbs) == dirAbs && questionNamePattern.MatchString(filepath.Base(sourceAbs)) {
		if _, err := os.Stat(sourceAbs); err != nil {
			return "", false, fmt.Errorf("open question %q: %w", sourcePath, err)
		}
		return sourceAbs, true, nil
	}

	extension := filepath.Ext(sourcePath)
	if extension != ".md" && extension != ".html" {
		return "", false, fmt.Errorf("question %q must have a .md or .html extension", sourcePath)
	}
	destination, err := nextQuestionPath(dirAbs, extension)
	if err != nil {
		return "", false, err
	}
	if err := os.Rename(sourceAbs, destination); err != nil {
		return "", false, fmt.Errorf("move question %q to %q: %w", sourcePath, destination, err)
	}
	return destination, false, nil
}

func nextQuestionPath(interviewDir, extension string) (string, error) {
	entries, err := os.ReadDir(interviewDir)
	if err != nil {
		return "", fmt.Errorf("read interview directory %q: %w", interviewDir, err)
	}
	highest := 0
	for _, entry := range entries {
		matches := questionNamePattern.FindStringSubmatch(entry.Name())
		if len(matches) == 0 {
			continue
		}
		number, err := strconv.Atoi(matches[1])
		if err != nil {
			return "", fmt.Errorf("parse question number %q: %w", matches[1], err)
		}
		if number > highest {
			highest = number
		}
	}
	if highest >= 9999 {
		return "", fmt.Errorf("interview directory %q has no four-digit question numbers remaining", interviewDir)
	}
	return filepath.Join(interviewDir, fmt.Sprintf("%04d%s", highest+1, extension)), nil
}

func appendQuestionAsked(runDir, questionPath string) error {
	file, err := os.OpenFile(filepath.Join(runDir, "timeline.jsonl"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open run timeline: %w", err)
	}
	event := map[string]any{
		"type":     "QuestionAsked",
		"question": questionPath,
		"ts":       time.Now().UTC().Format(time.RFC3339Nano),
	}
	encodeErr := json.NewEncoder(file).Encode(event)
	closeErr := file.Close()
	if encodeErr != nil {
		return fmt.Errorf("append run timeline: %w", encodeErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close run timeline: %w", closeErr)
	}
	return nil
}

func waitForAnswer(command *cobra.Command, answerPath string) ([]byte, error) {
	ticker := time.NewTicker(answerPollInterval)
	defer ticker.Stop()
	for {
		answer, err := os.ReadFile(answerPath)
		if err == nil && len(answer) > 0 {
			return answer, nil
		}
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("read answer %q: %w", answerPath, err)
		}
		select {
		case <-command.Context().Done():
			return nil, command.Context().Err()
		case <-ticker.C:
		}
	}
}

func validQuestionPath(path string) (string, error) {
	if !questionNamePattern.MatchString(filepath.Base(path)) {
		return "", fmt.Errorf("question path %q must name a numbered .md or .html question", path)
	}
	if _, err := os.Stat(path); err != nil {
		return "", fmt.Errorf("open question %q: %w", path, err)
	}
	return path, nil
}

func answerPathForQuestion(questionPath string) string {
	return strings.TrimSuffix(questionPath, filepath.Ext(questionPath)) + ".answer.md"
}

func readAnswerInput(reader io.Reader) ([]byte, error) {
	if file, ok := reader.(*os.File); ok {
		info, err := file.Stat()
		if err != nil {
			return nil, fmt.Errorf("inspect stdin: %w", err)
		}
		if info.Mode()&os.ModeCharDevice != 0 {
			return nil, fmt.Errorf("answer text is required as an argument or on stdin")
		}
	}
	answer, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read answer from stdin: %w", err)
	}
	return answer, nil
}

func writeExclusive(path string, contents []byte) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return fmt.Errorf("answer %q already exists; refusing to overwrite", path)
		}
		return fmt.Errorf("create answer %q: %w", path, err)
	}
	writeErr := func() error {
		if _, err := file.Write(contents); err != nil {
			return fmt.Errorf("write answer %q: %w", path, err)
		}
		if err := file.Close(); err != nil {
			return fmt.Errorf("close answer %q: %w", path, err)
		}
		return nil
	}()
	if writeErr != nil {
		_ = file.Close()
		_ = os.Remove(path)
		return writeErr
	}
	return nil
}

func pathForDisplay(path string) string {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	workingDir, err := os.Getwd()
	if err != nil {
		return absolute
	}
	relative, err := filepath.Rel(workingDir, absolute)
	if err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return relative
	}
	return absolute
}
