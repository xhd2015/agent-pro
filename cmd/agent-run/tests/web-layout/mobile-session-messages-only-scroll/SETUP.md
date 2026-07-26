# Scenario

**Feature**: session page fixed chrome — only message-list scrolls

```
seed overflow transcript → session page → document fixed → scroll message-list → chrome positions stable
```

## Preconditions

- `playwright-debug` on PATH.
- Seeded `fake-opencode/layout-scroll` with ≥15 messages so transcript overflows mobile viewport.
- Web server with explicit `--token test-token`.

## Steps

1. Seed `layout-scroll` session under `AGENT_RUN_HOME`.
2. Start `agent-run web` on a free port.
3. Open session route with token in localStorage.
4. Assert document/body do not scroll vertically; `message-list` overflows.
5. Scroll `message-list` up; assert `.top-bar`, `.session-header`, and composer Y positions unchanged (±2px).

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	requirePlaywright(t)

	req.Layout = "session-messages-only-scroll"
	req.WebTokenMode = "explicit"
	req.Token = "test-token"
	runner := "fake-opencode"
	sessionID := "layout-scroll"

	if err := seedLayoutScrollSession(t, req.Home, runner, sessionID, 18); err != nil {
		return err
	}

	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}

	sessionPath := "/sessions/" + runner + "/" + sessionID
	body := seedTokenInPage(req.Token) + `
await page.goto('` + req.BaseURL + sessionPath + `', { waitUntil: 'networkidle' });
` + waitForChatActive() + assertNoDocumentVerticalScroll() + assertMessageListOverflows() +
		assertChromeFixedWhileMessageListScrolls()

	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}
```