# Scenario

**Bug**: finished session with terminal_session_id must not repeat GET /terminal

```
finished grok-tty + terminal_session_id + live registry
  -> open chat page -> passive watch 8s -> terminal GET <= 1
```

## Preconditions

- `playwright-debug` on PATH.
- Session detail returns `terminal_session_id` on first load.

## Steps

1. Seed finished session with known terminal mapping and live registry.
2. Start `agent-run web`.
3. Register terminal poll monitor; open session page; wait 8s.
4. Assert terminal GET count ≤ 1.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	requirePlaywright(t)

	req.Scenario = "known-terminal-id-no-repeat-poll"
	req.WatchWindowMs = 8000

	seedFinishedKnownTerminalSession(t, req)

	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}

	body := initTerminalPollMonitor() + sessionBrowserScript(req, assertKnownTerminalNoRepeatPoll(1, req.WatchWindowMs))
	req.PlaywrightScript = body
	return nil
}
```