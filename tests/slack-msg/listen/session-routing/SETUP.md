# Scenario

**Feature**: session mode routes via agentrunbridge (+ SYSTEM.md / open inject)

```
inbound event -> thread: RunInteractiveOpen + SYSTEM.md + open inject
             -> stateless: Run+CaptureStdout (no SYSTEM.md)
  -> mock agent-run argv log (launch only; tty status not counted)
```

## Preconditions

- Default `--session-mode thread` unless leaf sets stateless.
- Session id format: `slack-{channel}-{thread_ts}` as `--session-id=…`.
- Thread: `run` + `--auto-send-or-resume` + `--new-terminal` + `--open` (not `--keep-tty`, not `send`).
- Thread SYSTEM.md under `$HOME/.agent-pro/slack-local-bot/sessions/<sessionID>/`.
- Stateless: `run` without `--open` / session-id; no SYSTEM.md.
- Leaves that assert SYSTEM.md paths set `req.HomeDir` for isolation.

## Steps

1. Start daemon with session mode flags.
2. Inject one or two events in same thread.
3. Assert launch argv matches open profile or stateless `run` (and SYSTEM.md when applicable).

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
