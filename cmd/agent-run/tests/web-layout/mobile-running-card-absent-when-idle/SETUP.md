# Scenario

**Feature**: running status card only appears while session is running

```
seed idle session -> web + token -> session route -> no agent-running-card in DOM
```

## Preconditions

- `playwright-debug` on PATH.
- `meta.status` is `idle` (not `running`).
- Same session route and auth as active chat tests.

## Steps

1. Seed `fake-opencode/layout-idle` with `idle` status.
2. Start web with explicit token.
3. Open session URL; assert running card absent; chat shell still loads.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	requirePlaywright(t)

	req.Layout = "running-absent"
	req.WebTokenMode = "explicit"
	req.Token = "test-token"
	runner := "fake-opencode"
	sessionID := "layout-idle"

	if err := seedIdleSessionForRunningCardNegative(t, req.Home, runner, sessionID); err != nil {
		return err
	}

	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}

	sessionPath := "/sessions/" + runner + "/" + sessionID
	body := seedTokenInPage(req.Token) + `
await page.goto('` + req.BaseURL + sessionPath + `', { waitUntil: 'networkidle' });
const chat = page.locator('[data-testid="chat-active"]');
await chat.waitFor({ state: 'visible', timeout: 15000 });
` + assertAgentRunningCardAbsent() + assertComposerPinnedBottom()

	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}
```