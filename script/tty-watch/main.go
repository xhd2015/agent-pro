package main

import (
	"fmt"
	"os"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		printHelp()
		return nil
	}
	switch args[0] {
	case serveSubcommand:
		return runServeSession(args[1:])
	case "run":
		return runRun(args[1:])
	case "list":
		return runList(args[1:])
	case "watch":
		return runWatch(args[1:])
	case "snapshot":
		return runSnapshot(args[1:])
	case "kill":
		return runKill(args[1:])
	case "send":
		return runSend(args[1:])
	case "-h", "--help", "help":
		printHelp()
		return nil
	default:
		return fmt.Errorf("unknown subcommand %q", args[0])
	}
}

func printHelp() {
	fmt.Print(`tty-watch — embedded PTY session manager

Usage:
  tty-watch run <command> [args...]  Start session and attach (writer)
  tty-watch list                     List sessions
  tty-watch watch <session-id>       Observe session (readonly)
  tty-watch snapshot <session-id>    Print sanitized scrollback
  tty-watch kill <session-id>        End session and remove registry
  tty-watch send <session-id> <msg>  Inject follow-up text into live session

Options:
  -h, --help                         Show help
`)
}