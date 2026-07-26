# Scenario

**Feature**: live session streaming uses SSE tail, not session-detail polling

```
init transport monitor -> create grok-tty session -> session page streams 8s -> SSE >=1, detail GET <=3
```

## Preconditions

- Web started with grok mock harness (`--agent-runner grok-tty`).
- `playwright-debug` on PATH.

## Steps

1. Build grok mock env; start `agent-run web` on free port.
2. Register network monitor before navigation.
3. POST `grok-tty` session; open session page; wait 8s.
4. Assert SSE used and session-detail GET count bounded.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	requirePlaywright(t)

	req.Layout = "sse-transport"
	req.WebTokenMode = "omit"
	prompt := "monitor sse transport"
	marker := layoutGrokAssistantMarker(prompt)

	req.Port = findFreePort(t)
	if err := ensureLayoutGrokMockEnv(t, req, prompt, marker, 10); err != nil {
		return err
	}
	if err := startWebBackground(t, req); err != nil {
		return err
	}

	body := initStreamingTransportMonitor() +
		openLiveGrokTTYSession(req.BaseURL, prompt) +
		assertStreamingTransportProfile()

	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}
```