//go:build integration

package workflows_test

import (
	"debug/buildinfo"
	"fmt"
	"slices"
)

type canonicalValidation struct {
	Item     string `json:"item"`
	ExitCode int    `json:"exit_code"`
	Passed   bool   `json:"passed"`
	Infer    *struct {
		Verdict string `json:"verdict"`
	} `json:"infer,omitempty"`
}

type canonicalEvent struct {
	Type        string                `json:"type"`
	Name        string                `json:"name"`
	Item        string                `json:"item"`
	Next        string                `json:"next"`
	Verdict     string                `json:"verdict"`
	Validations []canonicalValidation `json:"validations"`
}

func verifyCanonicalTrace(events []canonicalEvent) error {
	want := []string{"Standard shipping", "Expedited shipping"}
	var selected, verdicts []string
	var validations [][]canonicalValidation
	implemented, reviewed, completed := false, false, false
	for _, event := range events {
		switch event.Type {
		case "LoopItemSelected":
			if len(selected) == 0 || selected[len(selected)-1] != event.Item {
				selected = append(selected, event.Item)
			}
			implemented, reviewed = false, false
		case "StageCompleted":
			switch event.Name {
			case "implement":
				implemented = true
			case "review":
				reviewed = implemented && event.Next == "sprints"
			}
		case "LoopValidated":
			if !reviewed {
				return fmt.Errorf("validation without implementation and review returning to the loop")
			}
			validations = append(validations, event.Validations)
		case "LoopEvaluated":
			verdicts = append(verdicts, event.Verdict)
		case "PipelineCompleted":
			completed = true
		}
	}
	if !completed {
		return fmt.Errorf("pipeline did not complete")
	}
	if !slices.Equal(selected, want) {
		return fmt.Errorf("item dispatch order: got %v, want %v", selected, want)
	}
	if len(validations) < 2 || len(validations[0]) != 1 || validations[0][0].Item != want[0] {
		return fmt.Errorf("first sprint was not validated separately")
	}
	final := validations[len(validations)-1]
	if len(final) != 2 || final[0].Item != want[0] || final[1].Item != want[1] {
		return fmt.Errorf("final lap did not revalidate both items")
	}
	for _, result := range final {
		if !result.Passed || result.ExitCode != 0 || result.Infer == nil || result.Infer.Verdict != "pass" {
			return fmt.Errorf("command and evidence judge did not pass %s", result.Item)
		}
	}
	if len(verdicts) < 2 || verdicts[0] != "not_done" || verdicts[len(verdicts)-1] != "done" {
		return fmt.Errorf("goal evaluator did not distinguish partial from complete work: %v", verdicts)
	}
	return nil
}

func verifyCanonicalBuild(binary, revision string) error {
	info, err := buildinfo.ReadFile(binary)
	if err != nil {
		return fmt.Errorf("read candidate Go build metadata: %w", err)
	}
	var builtRevision, modified string
	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			builtRevision = setting.Value
		case "vcs.modified":
			modified = setting.Value
		}
	}
	if builtRevision != revision || modified != "false" {
		return fmt.Errorf("candidate must come from clean commit %s; got revision=%q modified=%q", revision, builtRevision, modified)
	}
	return nil
}
