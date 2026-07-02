# Scenario

**Feature**: follow mode — auto-scroll while user is at bottom during streaming

```
live fake-codex stream → message-list at bottom → assistant text grows → scrollTop stays at max
```

## Preconditions

- `fake-codex` built into temp `bin/` and on PATH.
- Open API (`WebTokenMode=omit`).
- Live session created via POST; assistant bubble text grows during SSE stream.

## Steps

1. Build `fake-codex`, start web with open API.
2. Create live session and open session route.
3. Ensure `message-list` starts at bottom (`distanceFromBottom <= 80`).
4. Poll while assistant text grows; assert `scrollTop` remains at effective bottom on each growth tick.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	requirePlaywright(t)

	if err := buildFakeCodexIntoPath(t, req); err != nil {
		return err
	}

	req.Layout = "session-auto-follow-at-bottom"
	req.WebTokenMode = "omit"
	req.Port = findFreePort(t)
	if err := startWebBackground(t, req); err != nil {
		return err
	}

	body := openLiveFakeCodexSession(req.BaseURL, "grow the stream for follow mode") +
		assertMessageListAtBottom() + assertFollowAtBottomDuringStreaming()

	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}
```