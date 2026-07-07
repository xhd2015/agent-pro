# Scenario

**Feature**: assistant reply streams into a single bubble with growing visible text

```
POST grok-tty create on session page -> assistant events -> message-item-assistant text length increases
```

## Preconditions

- `llm-mock-run-grok` hook seeds deterministic assistant text under `--grok-home`.
- Open API or token in localStorage; live run emits streaming assistant content.

## Steps

1. Start web with grok mock flags (`WebTokenMode=omit`) on free port.
2. Navigate to home, create `grok-tty` session via API from playwright.
3. Poll assistant bubble text until length increases.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	requirePlaywright(t)

	req.Layout = "streaming-bubble"
	req.WebTokenMode = "omit"
	prompt := "grow the stream"
	marker := layoutGrokStreamMarker

	req.Port = findFreePort(t)
	if err := ensureLayoutGrokMockEnv(t, req, prompt, marker, 8); err != nil {
		return err
	}
	if err := startWebBackground(t, req); err != nil {
		return err
	}

	createJS := openLiveGrokTTYSession(req.BaseURL, prompt)
	body := createJS + assertAssistantBubbleTextGrows()

	req.PlaywrightScript = mobileViewportScript(body)
	return nil
}
```