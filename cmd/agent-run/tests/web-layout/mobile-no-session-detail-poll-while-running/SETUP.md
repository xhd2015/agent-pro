# Scenario

**Feature**: running live session does not meta-poll session detail while SSE is active

```
init detail-poll monitor -> create grok-tty session -> passive watch 15s -> detail GET === 1, SSE === 1, aborted === 0
```

## Preconditions

- Web started with grok mock harness (`--agent-runner grok-tty`).
- Mock hook sleeps long enough to keep session `running` through the watch window.

## Steps

1. Build grok mock env; start `agent-run web` on free port.
2. Register detail-poll monitor before navigation.
3. POST `grok-tty` session; open session page; passive wait 15s.
4. Assert exactly one detail GET and one SSE stream with zero aborts.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	requirePlaywright(t)

	req.Layout = "no-detail-poll"
	req.WebTokenMode = "omit"
	prompt := "passive watch no detail poll"
	marker := layoutGrokAssistantMarker(prompt)

	req.Port = findFreePort(t)
	if err := ensureLayoutGrokMockEnv(t, req, prompt, marker, 18); err != nil {
		return err
	}
	if err := startWebBackground(t, req); err != nil {
		return err
	}

	body := initSessionDetailPollMonitor() +
		openLiveGrokTTYSession(req.BaseURL, prompt) +
		assertNoSessionDetailPollWhileRunning(1, 15000)

	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}
```