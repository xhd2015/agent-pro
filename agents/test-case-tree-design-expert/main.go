package main

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/xhd2015/agent-pro/agent/agentui"
)

//go:embed PROMPT.md
var prompt string

const usage = `Usage: test-case-tree-design-expert [--model MODEL] [--agent-runner RUNNER] [--output DIR] [--resume SESSION_ID] <feature description>

Generate a tree-expanded test case design for the given feature description.
Produces a directory with SETUP.md (inherited across branches), ASSERT.md (per leaf),
and a README.md overview with mermaid graph, text tree, and test case index.

Arguments:
  <feature description>   The feature to design tests for (positional args joined with spaces)

Options:
  --model MODEL           Model identifier to use (defaults to the first free model from opencode)
  --agent-runner RUNNER   Agent runner to use (opencode or codex, default: opencode)
  -o, --output DIR        Directory to write the test case tree (default: auto-generated <name>-test-cases/)
  --resume SESSION_ID     Resume a previous session by its ID
  -h, --help              Show this help message
`

func main() {
	if err := agentui.Run(agentui.Config{
		AgentName:     "test-case-tree-design-expert",
		SessionPrefix: "tctd_",
		Prompt:        prompt,
		Usage:         usage,
	}, os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
