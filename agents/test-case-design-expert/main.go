package main

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/xhd2015/agent-pro/agent/agentui"
)

//go:embed PROMPT.md
var prompt string

const usage = `Usage: test-case-design-expert [--model MODEL] [--output FILE] [--resume SESSION_ID] <feature description>

Generate user-facing end-to-end test cases for the given feature description.
The expert will brainstorm the idea, ask clarifying questions interactively,
and produce a complete test plan.

Arguments:
  <feature description>   The feature to design tests for (positional args joined with spaces)

Options:
  --model MODEL      Model identifier to use (defaults to the first free model from opencode)
  -o, --output FILE   File to write the test report (default: auto-generated <name>-tests-design.md)
  --resume SESSION_ID Resume a previous session by its ID
  -h, --help         Show this help message
`

func main() {
	if err := agentui.Run(agentui.Config{
		AgentName:     "test-case-design-expert",
		SessionPrefix: "tcd_",
		Prompt:        prompt,
		Usage:         usage,
		OutputSuffix:  "-tests-design.md",
	}, os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
