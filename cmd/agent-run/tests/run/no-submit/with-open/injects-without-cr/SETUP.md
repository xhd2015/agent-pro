# Scenario

**Feature**: `--open --no-submit "draft"` injects without trailing Enter — no auto-submit

```
agent-run run --agent-runner grok-tty --open --no-submit "draft"
  -> CR-sensitive fake TUI: no SUBMITTED:draft in PTY snapshot
  -> (instant attach returns)
  -> stderr contains "grok-tty: <session-id>" after attach
  -> open lifecycle completes (exit 0)
```

## Preconditions

- Fake TUI uses `fakeTUIRequiresCR()` — prints `SUBMITTED:<line>` only after
  Enter completes the line. Bare type-without-CR must not produce `SUBMITTED:`.
- `Mode=open-snapshot-after` loads registry and snapshots PTY text after open.
- `AGENT_RUN_OPEN_ATTACH_INSTANT=1` so the CLI process exits after inject.

## Steps

1. Run open + no-submit with draft prompt and CR-only fake TUI.
2. After exit 0, snapshot PTY scrollback via registry listen_addr.
3. Assert no `SUBMITTED:draft`; session id printed once after attach.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Prompt = "draft"
	req.Mode = "open-snapshot-after"
	req.OpenInstantAttach = true
	setGrokTTYCommand(req, fakeTUIRequiresCR())
	req.Args = []string{
		"run",
		"--agent-runner", "grok-tty",
		"--open",
		"--no-submit",
		req.Prompt,
	}
	req.ExecTimeout = 60 * time.Second
	return nil
}
```
