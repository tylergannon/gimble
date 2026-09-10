package main

import (
	"fmt"
	"strings"
)

type Input struct {
	Goal       string     `json:"goal"`
	Acceptance Acceptance `json:"acceptance"`
}

type Acceptance struct {
	Command string `json:"command"`
}

func (input Input) validate() error {
	if strings.TrimSpace(input.Goal) == "" || strings.TrimSpace(input.Acceptance.Command) == "" {
		return fmt.Errorf("goal and acceptance.command are required")
	}
	return nil
}
