# Scenario

**Feature**: home detached follow mode — sending a message does not auto-scroll

```
seed 20 home sessions → scroll up to detach → composer send (new grok-tty session) → scrollTop unchanged before navigation
```

## Preconditions

- Web started with grok mock harness for composer-created `grok-tty` sessions.
- Open API.
- Seeded 20 home sessions with overflow list (runner label is fixture-only).

## Steps

1. Seed 20 home sessions under `fake-codex` (static fixture runner).
2. Start grok mock web with open API; open `/`.
3. Scroll `session-list` up to detach; record `scrollTop`.
4. Send message via composer; within 500ms assert `scrollTop` unchanged (±2px).

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	requirePlaywright(t)

	req.Layout = "home-send-no-auto-scroll"
	req.WebTokenMode = "omit"
	runner := "fake-codex"
	prompt := "home follow-up while detached"
	marker := layoutGrokAssistantMarker(prompt)

	if err := seedManyHomeSessions(t, req.Home, runner, 20); err != nil {
		return err
	}

	req.Port = findFreePort(t)
	if err := ensureLayoutGrokMockEnv(t, req, prompt, marker, 6); err != nil {
		return err
	}
	if err := startWebBackground(t, req); err != nil {
		return err
	}

	body := openHomePage(req.BaseURL) + waitForSessionListOverflow() +
		scrollSessionListUpFromBottom(250) + assertSessionListDetached() +
		recordSessionListScrollTop("BeforeSend") + sendComposerMessage(prompt) +
		`
await page.waitForTimeout(500);
` + assertSessionListScrollTopEqualsVar("BeforeSend")

	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}
```