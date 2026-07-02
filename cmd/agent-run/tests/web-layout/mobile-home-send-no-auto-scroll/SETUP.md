# Scenario

**Feature**: home detached follow mode — sending a message does not auto-scroll

```
seed 20 home sessions → scroll up to detach → composer send → scrollTop unchanged before navigation
```

## Preconditions

- `fake-codex` built into temp `bin/` and on PATH.
- Open API.
- Seeded 20 home sessions with overflow list.

## Steps

1. Seed 20 home sessions under `fake-codex`.
2. Start web with open API; open `/`.
3. Scroll `session-list` up to detach; record `scrollTop`.
4. Send message via composer; within 500ms assert `scrollTop` unchanged (±2px).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	requirePlaywright(t)

	if err := buildFakeCodexIntoPath(t, req); err != nil {
		return err
	}

	req.Layout = "home-send-no-auto-scroll"
	req.WebTokenMode = "omit"
	runner := "fake-codex"

	if err := seedManyHomeSessions(t, req.Home, runner, 20); err != nil {
		return err
	}

	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}

	body := openHomePage(req.BaseURL) + waitForSessionListOverflow() +
		scrollSessionListUpFromBottom(250) + assertSessionListDetached() +
		recordSessionListScrollTop("BeforeSend") + sendComposerMessage("home follow-up while detached") +
		`
await page.waitForTimeout(500);
` + assertSessionListScrollTopEqualsVar("BeforeSend")

	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}
```