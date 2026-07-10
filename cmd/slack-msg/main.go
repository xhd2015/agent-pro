package main

import (
	"fmt"
	"os"
	"strings"
)

const rootHelpText = `slack-msg: Slack messaging CLI.

Usage:
  slack-msg <command> [options]

Commands:
  send     Post a message via Slack Web API
  history  Fetch conversation history or thread replies
  listen   Socket Mode inbound bridge to agent-run

Options:
  -h, --help  Show help
`

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		fmt.Print(rootHelpText)
		return nil
	}

	switch args[0] {
	case "-h", "--help":
		fmt.Print(rootHelpText)
		return nil
	case "send":
		return runSend(args[1:])
	case "history":
		return runHistory(args[1:])
	case "listen":
		return runListenCommand(args[1:])
	default:
		if strings.HasPrefix(args[0], "-") {
			return fmt.Errorf("unknown option: %s", args[0])
		}
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

func flagString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
