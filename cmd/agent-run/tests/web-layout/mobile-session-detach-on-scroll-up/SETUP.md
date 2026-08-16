# Scenario

**Feature**: follow mode — detach when user scrolls up during streaming

```
seeded overflow grok-tty session → scroll up to detach → composer follow-up stream → scrollTop frozen
```

## Preconditions

- Web started with grok mock harness.
- Open API.
- Seeded idle session with overflow history so detach scroll-up is meaningful before follow-up stream.

## Steps

1. Seed `grok-tty/layout-scroll-detach` with ≥15 messages.
2. Open session route; wait for `message-list` overflow.
3. Scroll `message-list` up ≥200px from bottom; record `scrollTop`.
4. Send composer follow-up to start live stream.
5. Poll while assistant text grows; assert `scrollTop` stable (±2px).

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	requirePlaywright(t)

	req.Layout = "session-detach-on-scroll-up"
	req.WebTokenMode = "omit"
	runner := "grok-tty"
	sessionID := "layout-scroll-detach"
	followUpPrompt := "grow the stream for detach mode"
	marker := layoutGrokStreamMarker

	if err := seedLayoutScrollSession(t, req.Home, runner, sessionID, 18); err != nil {
		return err
	}

	req.Port = findFreePort(t)
	if err := ensureLayoutGrokMockEnv(t, req, followUpPrompt, marker, 10); err != nil {
		return err
	}
	if err := startWebBackground(t, req); err != nil {
		return err
	}

	sessionPath := "/sessions/" + sessionID
	body := openSeededSessionPage(req.BaseURL, sessionPath) +
		waitForMessageListOverflow() + scrollMessageListUpFromBottom(250) +
		assertMessageListDetached() + recordFrozenScrollTop() +
		sendComposerMessage(followUpPrompt) +
		assertScrollTopFrozenDuringStreaming()

	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}
```