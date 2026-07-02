# Scenario

**Feature**: follow mode — detach when user scrolls up during streaming

```
seeded overflow session → scroll up to detach → composer follow-up stream → scrollTop frozen
```

## Preconditions

- `fake-codex` on PATH.
- Open API.
- Seeded idle session with overflow history so detach scroll-up is meaningful before follow-up stream.

## Steps

1. Seed `fake-codex/layout-scroll-detach` with ≥15 messages.
2. Open session route; wait for `message-list` overflow.
3. Scroll `message-list` up ≥200px from bottom; record `scrollTop`.
4. Send composer follow-up to start live stream.
5. Poll while assistant text grows; assert `scrollTop` stable (±2px).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	requirePlaywright(t)

	if err := buildFakeCodexIntoPath(t, req); err != nil {
		return err
	}

	req.Layout = "session-detach-on-scroll-up"
	req.WebTokenMode = "omit"
	runner := "fake-codex"
	sessionID := "layout-scroll-detach"

	if err := seedLayoutScrollSession(t, req.Home, runner, sessionID, 18); err != nil {
		return err
	}

	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}

	sessionPath := "/sessions/" + runner + "/" + sessionID
	body := openSeededSessionPage(req.BaseURL, sessionPath) +
		waitForMessageListOverflow() + scrollMessageListUpFromBottom(250) +
		assertMessageListDetached() + recordFrozenScrollTop() +
		sendComposerMessage("grow the stream for detach mode") +
		assertScrollTopFrozenDuringStreaming()

	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}
```