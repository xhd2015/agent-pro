package main

const attachHelp = `
Usage: agent-run attach <session-id>

Attach to a live grok-tty or codex-tty session by registry id (printed on stderr during run).

Options:
  -h, --help          show help
`

func runAttach(args []string) error {
	return runTtyAttach(args)
}
