# Scenario

**Feature**: detached follow mode — sending a message does not auto-scroll

```
seed overflow idle session → scroll up to detach → composer send → scrollTop unchanged
```

## Preconditions

- `fake-codex` on PATH (composer follow-up triggers a run).
- Open API.
- Seeded idle session with overflow history (≥15 messages).

## Steps

1. Seed `fake-codex/layout-scroll-send` with many messages.
2. Start web, open session route.
3. Scroll `message-list` up to detach; record `scrollTop`.
4. Send follow-up via composer; assert `scrollTop` unchanged (±2px).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	requirePlaywright(t)

	if err := buildFakeCodexIntoPath(t, req); err != nil {
		return err
	}

	req.Layout = "session-send-no-auto-scroll"
	req.WebTokenMode = "omit"
	runner := "fake-codex"
	sessionID := "layout-scroll-send"

	if err := seedLayoutScrollSession(t, req.Home, runner, sessionID, 18); err != nil {
		return err
	}

	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}

	sessionPath := "/sessions/" + runner + "/" + sessionID
	body := `
await page.goto('` + req.BaseURL + sessionPath + `', { waitUntil: 'domcontentloaded' });
` + waitForChatActive() + waitForMessageListOverflow() + scrollMessageListUpFromBottom(250) + assertMessageListDetached() +
		recordMessageListScrollTop("BeforeSend") + sendComposerMessage("follow-up while detached") +
		`
await page.waitForTimeout(500);
` + assertMessageListScrollTopEqualsVar("BeforeSend")

	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}
```