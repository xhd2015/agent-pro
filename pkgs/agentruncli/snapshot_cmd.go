package agentruncli

const snapshotHelp = `
Usage: agent-run snapshot <session-id>

Print a sanitized snapshot of a live grok-tty or codex-tty session by registry id.

Options:
  -h, --help   show help
`

func runSnapshot(args []string) error {
	return runTtySnapshot(args)
}