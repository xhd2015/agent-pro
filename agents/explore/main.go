package main

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/xhd2015/agent-pro/agents/explore/run"
	"github.com/xhd2015/skills/install"
)

//go:embed run/SKILL.md
var skillContent string

const skillName = "explore"

const help = `
Usage: explore [OPTIONS] <prompt>

A file search specialist for exploring codebases. Uses glob patterns, regex
searches, and file reads to answer questions about the codebase. Delegates
to the parent agent runner (auto-detected) for LLM calls.

Options:
  --agent-runner <id>      override agent runner (opencode|pi|codex|crush)
  --model <model>          override model
  --session-id <id>        resume an existing session
  --timeout <duration>     timeout (default: 1h, min: 1m, e.g. "30m", "2h")
  --catch-up               replay session events (requires --session-id)
  --status                 show session status (requires --session-id)
  --list-sessions          list all explore sessions
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
				fmt.Fprintf(os.Stderr, "explore: %v\n", err)
				os.Exit(1)
			}
			return
		}
	}
	if err := run.Run(args); err != nil {
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
			Usage:        "explore skill install",
		}, args[1:])
	default:
		return fmt.Errorf("unknown skill sub-command: %s, expected skill show or skill install", args[0])
	}
}
