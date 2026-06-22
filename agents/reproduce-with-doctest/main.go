package main

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/xhd2015/agent-pro/agents/reproduce-with-doctest/run"
	"github.com/xhd2015/skills/install"
)

//go:embed run/SKILL.md
var skillContent string

const skillName = "reproduce-with-doctest"

const help = `
Usage: reproduce-with-doctest [OPTIONS] <prompt>

A doctest-backed bug reproduction specialist. Reproduces bugs by adding failing
doctest cases to existing test trees. Requires at least one RED doctest run.
Delegates to the parent agent runner (auto-detected) for LLM calls.

Options:
  --agent-runner <id>      override agent runner (opencode|pi|codex|crush|grok)
  --model <model>          override model
  --model-env <env>        override the env var used to pass the model
  --session-id <id>        resume an existing session
  --timeout <duration>     timeout (default: 1h, min: 1m, e.g. "30m", "2h")
  --catch-up               replay session events (requires --session-id)
  --status                 show session status (requires --session-id)
  --list-sessions          list all reproduce-with-doctest sessions
  --session-base <dir>     override sessions directory
  -h, --help               show this help

Skill Commands:
  skill show               print the embedded SKILL.md content
  skill install [OPTIONS] [DIR]
                           install the skill to .agents/skills, .codex/skills,
                           .cursor/skills, or .opencode/skills
`

func main() {
	args := os.Args[1:]
	if len(args) > 0 {
		switch args[0] {
		case "-h", "--help":
			fmt.Print(help)
			return
		case "skill":
			if err := handleSkillCommand(args[1:]); err != nil {
				fmt.Fprintf(os.Stderr, "reproduce-with-doctest: %v\n", err)
				os.Exit(1)
			}
			return
		}
	}
	if err := run.RunArgs(args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func handleSkillCommand(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("expected skill show or skill install")
	}
	switch args[0] {
	case "show":
		if len(args) > 1 {
			return fmt.Errorf("unexpected arguments after show")
		}
		fmt.Print(skillContent)
		return nil
	case "install":
		return install.HandleInstall(install.InstallOptions{
			SkillDirName: skillName,
			SkillContent: skillContent,
			Usage:        "reproduce-with-doctest skill install",
		}, args[1:])
	default:
		return fmt.Errorf("unknown skill sub-command: %s, expected skill show or skill install", args[0])
	}
}