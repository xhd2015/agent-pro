package main

import (
	"testing"

	"github.com/xhd2015/agent-pro/agent/cli/registry"
	"github.com/xhd2015/agent-pro/pkgs/agentui"
)

func TestExpandHomePath(t *testing.T) {
	home, err := expandHomePath("~/.mock/grok")
	if err != nil {
		t.Fatal(err)
	}
	if home == "" || home == "~/.mock/grok" {
		t.Fatalf("expandHomePath = %q", home)
	}
}

func TestApplyWebGrokRunOptions(t *testing.T) {
	cfg := webRunConfig{
		GrokHome:            "/mock/grok",
		GrokTTYRunnerBinary: "llm-mock-run-grok",
	}
	opts := &agentui.RunOptions{}
	applyWebGrokRunOptions(string(registry.AgentRunnerGrokTTY), cfg, opts)
	if opts.AgentRunnerConfigHome != "/mock/grok" {
		t.Fatalf("AgentRunnerConfigHome = %q", opts.AgentRunnerConfigHome)
	}
	if opts.AgentRunnerBinary != "llm-mock-run-grok" {
		t.Fatalf("AgentRunnerBinary = %q", opts.AgentRunnerBinary)
	}
}