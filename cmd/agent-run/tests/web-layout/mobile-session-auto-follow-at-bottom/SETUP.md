# Scenario

**Feature**: follow mode — auto-scroll while user is at bottom during streaming

```
live grok-tty stream → message-list at bottom → assistant text grows → scrollTop stays at max
```

## Preconditions

- Web started with grok mock harness.
- Open API.

## Steps

1. Start grok mock web on free port.
2. POST `grok-tty` session and open session page at bottom.
3. Poll while assistant text grows; assert `distanceFromBottom <= 80`.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	requirePlaywright(t)

	req.Layout = "session-auto-follow-at-bottom"
	req.WebTokenMode = "omit"
	prompt := "grow the stream for follow mode"
	marker := layoutGrokStreamMarker

	req.Port = findFreePort(t)
	if err := ensureLayoutGrokMockEnv(t, req, prompt, marker, 10); err != nil {
		return err
	}
	if err := startWebBackground(t, req); err != nil {
		return err
	}

	body := openLiveGrokTTYSession(req.BaseURL, prompt) +
		assertFollowAtBottomDuringStreaming()

	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}
```