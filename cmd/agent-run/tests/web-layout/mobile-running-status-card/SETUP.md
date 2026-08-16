# Scenario

**Bug**: session page lacks prominent running indicator with elapsed duration

```
seed running session(updated_at ~30s ago) -> web + token -> /sessions/... -> agent-running-card + duration
```

## Preconditions

- `playwright-debug` on PATH (else `t.Skip`).
- `meta.status` is `running` with `updated_at` about 30 seconds before now.
- Web server uses explicit `--token test-token`.

## Steps

1. Seed `fake-opencode/layout-running` with `running` status and one assistant message event.
2. Start `agent-run web` on a free port.
3. Open session route with token in localStorage; assert running card and duration text.

```go
import (
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	requirePlaywright(t)

	req.Layout = "running-card"
	req.WebTokenMode = "explicit"
	req.Token = "test-token"
	runner := "fake-opencode"
	sessionID := "layout-running"

	if err := seedRunningSession(t, req.Home, runner, sessionID, 30*time.Second); err != nil {
		return err
	}

	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}

	sessionPath := "/sessions/" + sessionID
	body := seedTokenInPage(req.Token) + `
await page.goto('` + req.BaseURL + sessionPath + `', { waitUntil: 'networkidle' });
` + assertAgentRunningCardVisible() + assertComposerPinnedBottom()

	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}
```