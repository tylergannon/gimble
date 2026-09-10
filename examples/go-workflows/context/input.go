package main

import (
	"fmt"
	"strings"
)

type Input struct {
	Goal string `json:"goal"`
}

func (input Input) validate() error {
	if strings.TrimSpace(input.Goal) == "" {
		return fmt.Errorf("goal is required")
	}
	return nil
}
