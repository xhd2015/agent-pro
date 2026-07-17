package agentruncli

import (
	"testing"

	"github.com/xhd2015/agent-pro/agent/cli/registry"
)

func TestGrokTTYUserMessageFromTail(t *testing.T) {
	if !grokTTYUserMessageFromTail(string(registry.AgentRunnerGrokTTY)) {
		t.Fatal("expected grok-tty user messages from tail only")
	}
}