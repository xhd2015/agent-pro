package main

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/xhd2015/agent-pro/agent/agentui"
)

//go:embed PROMPT.md
var prompt string

const usage = `Usage: idea-expander [--model MODEL] [--output FILE] [--resume SESSION_ID] <feature idea>

Expand a feature idea into a fully described plan with user stories, workflows,
and implementation considerations. The expert will brainstorm the idea, ask
clarifying questions interactively, and produce a complete expansion.

Arguments:
  <feature idea>   The idea to expand (positional args joined with spaces)

Options:
  --model MODEL      Model identifier to use (defaults to the first free model from opencode)
  -o, --output FILE   File to write the expansion report (default: auto-generated <name>-idea.md)
  --resume SESSION_ID Resume a previous session by its ID
  -h, --help         Show this help message
`

func main() {
	if err := agentui.Run(agentui.Config{
		AgentName:     "idea-expander",
		SessionPrefix: "ie_",
		Prompt:        prompt,
		Usage:         usage,
		OutputSuffix:  "-idea.md",
	}, os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
