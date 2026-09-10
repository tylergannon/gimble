package main

import (
	"fmt"
	"strings"
)

type Input struct {
	Topic string `json:"topic"`
}

func (input Input) validate() error {
	if strings.TrimSpace(input.Topic) == "" {
		return fmt.Errorf("topic is required")
	}
	return nil
}
