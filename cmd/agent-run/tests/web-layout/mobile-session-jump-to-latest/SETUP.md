# Scenario

**Feature**: jump-to-latest chip restores follow after detach

```
seeded overflow session → scroll up to detach → follow-up stream → chip visible → tap → bottom
```

## Preconditions

- `fake-codex` on PATH.
- Open API.
- Seeded overflow history so scroll-up detach works before follow-up stream adds content below.

## Steps

1. Seed `fake-codex/layout-scroll-jump` with ≥15 messages.
2. Open session route; wait for `message-list` overflow.
3. Scroll `message-list` up ≥200px to detach.
4. Send composer follow-up to start streaming growth below viewport.
5. Wait for `[data-testid="jump-to-latest"]`; tap; assert bottom + chip hidden.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	requirePlaywright(t)

	if err := buildFakeCodexIntoPath(t, req); err != nil {
		return err
	}

	req.Layout = "session-jump-to-latest"
	req.WebTokenMode = "omit"
	runner := "fake-codex"
	sessionID := "layout-scroll-jump"

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
		assertMessageListDetached() + sendComposerMessage("grow the stream for jump chip") +
		assertJumpToLatestChipFlow()

	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}
```