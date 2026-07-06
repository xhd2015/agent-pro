package main

const watchHelp = `
Usage: agent-run watch <session-id>

Stream readonly output from a live grok-tty or codex-tty session by registry id.

Options:
  -h, --help   show help
`

func runWatch(args []string) error {
	return runTtyWatch(args)
}