package fakeagent

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Script struct {
	Events   []Event `json:"events"`
	ExitCode int     `json:"exit_code"`
	Stderr   string  `json:"stderr"`
	DelayMS  int     `json:"delay_ms"`
}

func LoadScript(path string) (*Script, error) {
	if strings.TrimSpace(path) == "" {
		return nil, fmt.Errorf("script path is empty")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read script %s: %w", path, err)
	}
	var script Script
	if err := json.Unmarshal(data, &script); err != nil {
		return nil, fmt.Errorf("parse script %s: %w", path, err)
	}
	if err := script.Validate(); err != nil {
		return nil, fmt.Errorf("invalid script %s: %w", path, err)
	}
	return &script, nil
}

func (s Script) Validate() error {
	if s.ExitCode < 0 {
		return fmt.Errorf("exit_code must be >= 0")
	}
	if s.DelayMS < 0 {
		return fmt.Errorf("delay_ms must be >= 0")
	}
	for i, event := range s.Events {
		if event.Type == "" {
			return fmt.Errorf("events[%d].type is required", i)
		}
		if event.Item != nil && event.Item.Type == "" {
			return fmt.Errorf("events[%d].item.type is required", i)
		}
	}
	return nil
}
