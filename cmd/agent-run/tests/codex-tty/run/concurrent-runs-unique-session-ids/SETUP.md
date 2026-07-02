# Scenario

**Bug**: concurrent `codex-tty` runs all publish `session-1`

```
agent-run run --agent-runner codex-tty "run ls" x 3 in the same AGENT_RUN_HOME
  -> each process prints codex-tty: session-N
  -> session ids must be unique across the shared registry/home
```

## Steps

1. Use a fake Codex TUI that stays alive briefly after reading the prompt.
2. Launch three `agent-run run --agent-runner codex-tty "run ls"` processes in parallel.
3. Collect each prefixed session id from stderr.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Mode = "concurrent-run-unique-ids"
	req.ConcurrentRuns = 3
	req.CodexTTYCommand = `sh -c 'printf "CODEX_TTY_BANNER\nCodex › "; read line; printf "Response: %s\n" "$line"; sleep 0.5'`
	req.Args = append(req.Args, "run ls")
	return nil
}
```
