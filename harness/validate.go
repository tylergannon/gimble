package harness

import (
	"errors"
	"fmt"
	"strings"
)

// ValidateCreateSessionInput validates the provider-neutral inputs available
// before a native session is created.
func ValidateCreateSessionInput(model, workdir string) error {
	if strings.TrimSpace(model) == "" {
		return errors.New("model must not be empty")
	}
	if strings.TrimSpace(workdir) == "" {
		return errors.New("workdir must not be empty")
	}
	return nil
}

// ValidateRunTurnInput validates a turn before native harness activity begins.
func ValidateRunTurnInput(input RunTurnInput, onEvent OnEvent) error {
	if strings.TrimSpace(input.SessionID) == "" {
		return errors.New("session ID must not be empty")
	}
	if err := ValidateCreateSessionInput(input.Model, input.Workdir); err != nil {
		return err
	}
	if strings.TrimSpace(input.ReasoningEffort) == "" {
		return errors.New("reasoning effort must not be empty")
	}
	if err := ValidateContentParts(input.Parts); err != nil {
		return err
	}
	if onEvent == nil {
		return errors.New("event callback must not be nil")
	}
	return nil
}

// ValidateSessionInput validates inputs for a session-scoped operation.
func ValidateSessionInput(sessionID, workdir string) error {
	if strings.TrimSpace(sessionID) == "" {
		return errors.New("session ID must not be empty")
	}
	if strings.TrimSpace(workdir) == "" {
		return errors.New("workdir must not be empty")
	}
	return nil
}

// ValidateContentParts enforces the non-empty ordered text-part contract.
func ValidateContentParts(parts []ContentPart) error {
	if len(parts) == 0 {
		return errors.New("content parts must not be empty")
	}
	for i, part := range parts {
		if part.Type != ContentPartText {
			return fmt.Errorf("content part %d has unsupported type %q", i, part.Type)
		}
	}
	return nil
}
