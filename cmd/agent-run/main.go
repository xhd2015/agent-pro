package main

import (
	"fmt"
	"os"
	"strings"
)

const help = `
Usage: agent-run [--agent-runner RUNNER] [--help]
       agent-run <command> [ARGS]

Commands:
  web        start localhost web UI and API
  run        headless one-shot agent invocation
  sessions   list stored sessions or print one session's events
  status     show agent-run status

Options:
  --agent-runner RUNNER   default agent runner (codex, opencode, fake-codex, ...)
  -h, --help              show help

Run agent-run <command> --help for command-specific options.
`

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "agent-run: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	var agentRunner string
	var cmdArgs []string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-h", "--help":
			if len(cmdArgs) == 0 {
				fmt.Print(strings.TrimPrefix(help, "\n"))
				return nil
			}
			cmdArgs = append(cmdArgs, args[i])
		case "--agent-runner":
			if i+1 >= len(args) {
				return fmt.Errorf("--agent-runner requires a value")
			}
			agentRunner = args[i+1]
			i++
		default:
			cmdArgs = append(cmdArgs, args[i])
		}
	}
	if len(cmdArgs) == 0 {
		fmt.Print(strings.TrimPrefix(help, "\n"))
		return nil
	}
	cmd := cmdArgs[0]
	sub := cmdArgs[1:]
	switch cmd {
	case "web":
		return runWeb(sub, agentRunner)
	case "run":
		return runHeadless(sub, agentRunner)
	case "sessions":
		return runSessions(sub)
	case "status":
		return runStatus(sub)
	default:
		return fmt.Errorf("unknown command: %s", cmd)
	}
}