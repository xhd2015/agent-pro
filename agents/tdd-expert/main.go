package main

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/xhd2015/agent-pro/agent/agentui"
)

//go:embed PROMPT.md
var prompt string

//go:embed SKILL.md
var skillTemplate string

const usage = `Usage: tdd-expert [--model MODEL] [--agent-runner RUNNER] [--output DIR] [--resume SESSION_ID] <test-case-tree-dir>

Convert a test case tree (output of test-case-tree-design-expert) into runnable
Go tests following Test-Driven Development. Generates a stub implementation
that returns "not implemented" errors so all tests start in RED state,
then runs the tests to verify.

Arguments:
  <test-case-tree-dir>   Path to the test case tree directory (must contain README.md)

Options:
  --model MODEL           Model identifier to use (defaults to the first free model from opencode)
  --agent-runner RUNNER   Agent runner to use (opencode or codex, default: opencode)
  -o, --output DIR        Directory to write the generated Go code (default: auto-derived from feature slug)
  --resume SESSION_ID     Resume a previous session by its ID
  -h, --help              Show this help message

Skill Commands:
  skill show              Print the embedded SKILL.md content
  skill install [OPTIONS] [DIR]
                          Install the skill to .agents/skills, .codex/skills,
                          .cursor/skills, or .opencode/skills
`

func main() {
	if err := agentui.Run(agentui.Config{
		AgentName:     "tdd-expert",
		SessionPrefix: "tdd_",
		Prompt:        prompt,
		Usage:         usage,
		SkillName:     "tdd-expert",
		SkillContent:  skillTemplate,
	}, os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
