# Scenario

**Bug**: discovery poll stops after terminal becomes available

```
running grok-tty without terminal_session_id in meta
  -> registry appears after 1s
  -> terminal GET > 0 during discovery, then stops; total <= 8 in 8s
```

## Preconditions

- `playwright-debug` on PATH.
- Registry file absent at page load; written after 1s delay.

## Steps

1. Start ptywrap; seed running session **without** `terminal_session_id`.
2. Start web; schedule delayed registry write (1s).
3. Register terminal poll monitor; open session page; wait 8s.
4. Assert discovery polls are bounded and stop after `available: true`.

```go
import (
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	requirePlaywright(t)

	req.Scenario = "discovery-poll-stops-after-available"
	req.Status = "running"
	req.Prompt = "discovery poll running prompt"
	req.WatchWindowMs = 8000
	req.RegistryDelay = 1 * time.Second

	listenAddr := startMappedPtywrap(t, req)
	writeSessionFixtureWithoutTerminalID(t, req)

	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}

	body := initTerminalPollMonitor() + sessionBrowserScript(req, assertDiscoveryPollStopsAfterAvailable(8, req.WatchWindowMs))
	req.PlaywrightScript = body
	_ = listenAddr
	return nil
}
```