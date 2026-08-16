# Scenario

**Feature**: jump-to-latest chip restores follow after detach

```
seeded overflow grok-tty session → scroll up to detach → follow-up stream → chip visible → tap → bottom
```

## Preconditions

- Web started with grok mock harness.
- Open API.
- Seeded overflow history so scroll-up detach works before follow-up stream adds content below.

## Steps

1. Seed `grok-tty/layout-scroll-jump` with ≥15 messages.
2. Open session route; wait for `message-list` overflow.
3. Scroll `message-list` up ≥200px to detach.
4. Send composer follow-up to start streaming growth below viewport.
5. Wait for `[data-testid="jump-to-latest"]`; tap; assert bottom + chip hidden.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	requirePlaywright(t)

	req.Layout = "session-jump-to-latest"
	req.WebTokenMode = "omit"
	runner := "grok-tty"
	sessionID := "layout-scroll-jump"
	followUpPrompt := "grow the stream for jump chip"
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
		assertMessageListDetached() + sendComposerMessage(followUpPrompt) +
		assertJumpToLatestChipFlow()

	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}
```