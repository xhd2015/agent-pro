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

const usage = `Usage: idea-expander [--model MODEL] [--agent-runner RUNNER] [--output FILE] [--resume SESSION_ID] <feature idea>

Expand a feature idea into a fully described plan with user stories, workflows,
and implementation considerations. The expert will brainstorm the idea, ask
clarifying questions interactively, and produce a complete expansion.

Arguments:
  <feature idea>   The idea to expand (positional args joined with spaces)

Options:
  --model MODEL           Model identifier to use (defaults to the first free model from opencode)
  --agent-runner RUNNER   Agent runner to use (opencode or codex, default: opencode)
  -o, --output FILE       File to write the expansion report (default: auto-generated <name>-idea.md)
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
		AgentName:     "idea-expander",
		SessionPrefix: "ie_",
		Prompt:        prompt,
		Usage:         usage,
		SkillName:     "idea-expander",
		SkillContent:  skillTemplate,
	}, os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
