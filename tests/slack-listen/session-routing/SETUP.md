# Scenario

**Feature**: session mode routes agent-run commands

```
inbound event -> thread/stateless routing -> mock agent-run argv log
```

## Preconditions

- Default `--session-mode thread` unless leaf sets stateless.
- Session id format: `slack-{channel}-{thread_ts}`.

## Steps

1. Start daemon with session mode flags.
2. Inject one or two events in same thread.
3. Assert agent argv contains `run --keep-tty --session` or `send`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ClearSlackEnv = true
	prependListenTokens(req)
	req.Daemon = true
	req.Args = []string{"--no-require-mention"}
	return nil
}
```