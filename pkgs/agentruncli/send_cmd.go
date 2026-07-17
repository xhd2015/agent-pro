package agentruncli

const sendHelp = `
Usage: agent-run send [OPTIONS] <session-id> "message"

Send a follow-up message to a live grok-tty or codex-tty session by registry id.
On successful enqueue, prints the session-local message id (msg_1, msg_2, …) to stdout.

Use agent-run msg status|cancel <session-id>/<message-id> to inspect or remove queued messages.

Options:
  --no-wait            enqueue and exit immediately without waiting for delivery
  --max-wait DURATION  enqueue, print id, then wait up to DURATION for delivery
  --no-submit          inject without trailing Enter (no auto-submit); stored on queue entry
  -h, --help           show help
`

func runSend(args []string) error {
	return runTtySend(args)
}