package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	fakeagent "github.com/xhd2015/agent-pro/pkgs/fake-agent"
	"github.com/xhd2015/less-gen/flags"
)

const help = `
Usage: fake-codex <command> [ARGS]

A fake codex CLI that mimics codex behavior without real LLM calls.

Commands:
  exec              execute a prompt (mimics codex exec)
  -h,--help         show help

Run fake-codex exec -h for exec-specific options.
`

const execHelp = `
Usage: fake-codex exec [OPTIONS] [PROMPT]

Mimics codex exec --json output with randomly generated events.

Options:
  --json                          output JSON events (default: human-friendly)
  --cd <dir>                      working directory (ignored, for compatibility)
  --sandbox <mode>                sandbox mode (ignored, for compatibility)
  --model <model>                 model name (ignored, for compatibility)
  --skip-git-repo-check           skip git check (ignored, for compatibility)
  --seed <int>                    random seed for deterministic output
  --delay <ms>                    delay in ms between output lines (default: 0)
  -h,--help                       show help
`

func main() {
	if err := handle(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "fake-codex: %v\n", err)
		os.Exit(1)
	}
}

func handle(args []string) error {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		fmt.Print(strings.TrimPrefix(help, "\n"))
		return nil
	}

	switch args[0] {
	case "exec":
		return handleExec(args[1:])
	default:
		return fmt.Errorf("unknown command: %s (only 'exec' is currently supported)", args[0])
	}
}

func handleExec(args []string) error {
	var dirFlag *string
	var sandboxFlag *string
	var modelFlag *string
	var jsonFlag *bool
	var skipGitFlag *bool
	var seedFlag *int64
	var delayFlag *int

	remaining, err := flags.String("--cd", &dirFlag).
		String("--sandbox", &sandboxFlag).
		String("--model", &modelFlag).
		Bool("--json", &jsonFlag).
		Bool("--skip-git-repo-check", &skipGitFlag).
		Int("--seed", &seedFlag).
		Int("--delay", &delayFlag).
		Help("-h,--help", execHelp).
		Parse(args)
	if err != nil {
		return err
	}

	prompt := strings.Join(remaining, " ")
	if prompt == "" {
		prompt = "help with the task"
	}

	var seed int64
	if seedFlag != nil {
		seed = *seedFlag
	} else {
		seed = time.Now().UnixNano()
	}

	var delay time.Duration
	if delayFlag != nil && *delayFlag > 0 {
		delay = time.Duration(*delayFlag) * time.Millisecond
	}

	gen := fakeagent.NewGenerator(seed)
	events := gen.GenerateSession(prompt)

	jsonOutput := jsonFlag != nil && *jsonFlag

	if jsonOutput {
		lines, err := fakeagent.FormatCodexEvents(events)
		if err != nil {
			return fmt.Errorf("format events: %w", err)
		}
		for i, line := range lines {
			fmt.Println(line)
			if delay > 0 && i < len(lines)-1 {
				time.Sleep(delay)
			}
		}
	} else {
		text := fakeagent.FormatCodexEventsText(events)
		if delay > 0 {
			lines := strings.Split(text, "\n")
			for _, line := range lines {
				fmt.Println(line)
				time.Sleep(delay)
			}
		} else {
			fmt.Print(text)
		}
	}

	return nil
}
