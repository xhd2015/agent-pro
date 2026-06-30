# Scenario

**Bug**: while agent is working, inline loading bubble must appear under the user's message (not only top card)

```
seed running + user event only -> session page -> message-item-assistant-loading below user
```

## Preconditions

- Session `status=running` with one user message and no assistant reply yet.
- SPA renders inline loading when running without active assistant stream.

## Steps

1. Seed `fake-opencode/inline-loading` session.
2. Start web; open session route with token.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	requirePlaywright(t)

	req.Layout = "inline-loading"
	req.WebTokenMode = "explicit"
	req.Token = "test-token"
	runner := "fake-opencode"
	sessionID := "inline-loading"

	if err := seedRunningSessionAwaitingAssistant(t, req.Home, runner, sessionID); err != nil {
		return err
	}

	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}

	sessionPath := "/sessions/" + runner + "/" + sessionID
	body := seedTokenInPage(req.Token) + `
await page.goto('` + req.BaseURL + sessionPath + `', { waitUntil: 'networkidle' });
` + assertInlineAssistantLoadingBelowUser()

	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}
```