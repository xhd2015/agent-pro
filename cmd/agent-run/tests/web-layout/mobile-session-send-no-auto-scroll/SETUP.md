# Scenario

**Feature**: detached follow mode — sending a message does not auto-scroll

```
seed overflow idle grok-tty session → scroll up to detach → composer send → scrollTop unchanged
```

## Preconditions

- Web started with grok mock harness (composer follow-up triggers a `grok-tty` run).
- Open API.
- Seeded idle session with overflow history (≥15 messages).

## Steps

1. Seed `grok-tty/layout-scroll-send` with many messages.
2. Start grok mock web, open session route.
3. Scroll `message-list` up to detach; record `scrollTop`.
4. Send follow-up via composer; assert `scrollTop` unchanged (±2px).

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	requirePlaywright(t)

	req.Layout = "session-send-no-auto-scroll"
	req.WebTokenMode = "omit"
	runner := "grok-tty"
	sessionID := "layout-scroll-send"
	followUpPrompt := "follow-up while detached"
	marker := layoutGrokAssistantMarker(followUpPrompt)

	if err := seedLayoutScrollSession(t, req.Home, runner, sessionID, 18); err != nil {
		return err
	}

	req.Port = findFreePort(t)
	if err := ensureLayoutGrokMockEnv(t, req, followUpPrompt, marker, 6); err != nil {
		return err
	}
	if err := startWebBackground(t, req); err != nil {
		return err
	}

	sessionPath := "/sessions/" + runner + "/" + sessionID
	body := `
await page.goto('` + req.BaseURL + sessionPath + `', { waitUntil: 'domcontentloaded' });
` + waitForChatActive() + waitForMessageListOverflow() + scrollMessageListUpFromBottom(250) + assertMessageListDetached() +
		recordMessageListScrollTop("BeforeSend") + sendComposerMessage(followUpPrompt) +
		`
await page.waitForTimeout(500);
` + assertMessageListScrollTopEqualsVar("BeforeSend")

	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}
```