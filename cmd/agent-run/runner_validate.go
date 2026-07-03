package main

import (
	"fmt"
	"strings"

	"github.com/xhd2015/agent-pro/agent/cli/registry"
	"github.com/xhd2015/agent-pro/pkgs/ttyrunner"
)

func validateRunner(runner string) error {
	runner = strings.TrimSpace(runner)
	if runner == "" {
		return nil
	}
	if ttyrunner.IsTTYRunner(runner) {
		return nil
	}
	switch registry.AgentRunnerID(runner) {
	case registry.AgentRunnerCodex,
		registry.AgentRunnerOpencode,
		registry.AgentRunnerCursor,
		registry.AgentRunnerFakeCodex,
		registry.AgentRunnerCrush,
		registry.AgentRunnerPi,
		registry.AgentRunnerGrok,
		registry.AgentRunnerGrokTTY,
		registry.AgentRunnerCodexTTY:
		return nil
	default:
		return fmt.Errorf("unknown agent runner: %s", runner)
	}
}
