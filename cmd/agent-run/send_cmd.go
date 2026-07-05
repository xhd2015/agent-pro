package main

const sendHelp = `
Usage: agent-run send <session-id> "message"

Send a follow-up message to a live grok-tty or codex-tty session by registry id.

Options:
  -h, --help   show help
`

func runSend(args []string) error {
	return runTtySend(args)
}