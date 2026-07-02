# Scenario

**Feature**: home page fixed chrome — only session-list scrolls

```
seed ≥20 home sessions → / → document fixed → scroll session-list → top-bar-home + composer stable
```

## Preconditions

- `playwright-debug` on PATH.
- Seeded `fake-codex` runner with ≥20 sessions so `session-list` overflows mobile viewport.
- Web server with explicit `--token test-token`.

## Steps

1. Seed 20 home sessions under `AGENT_RUN_HOME`.
2. Start `agent-run web` on a free port.
3. Open `/` with token in localStorage.
4. Assert document/body do not scroll vertically; `session-list` overflows.
5. Scroll `session-list` up; assert `.top-bar-home` and composer Y positions unchanged (±2px).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	requirePlaywright(t)

	req.Layout = "home-sessions-only-scroll"
	req.WebTokenMode = "explicit"
	req.Token = "test-token"
	runner := "fake-codex"

	if err := seedManyHomeSessions(t, req.Home, runner, 20); err != nil {
		return err
	}

	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}

	body := seedTokenInPage(req.Token) + openHomePage(req.BaseURL) +
		assertNoDocumentVerticalScroll() + assertSessionListOverflows() +
		assertChromeFixedWhileSessionListScrolls()

	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}
```