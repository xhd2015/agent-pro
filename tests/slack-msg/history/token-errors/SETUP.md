# Scenario

**Feature**: reject missing bot token for history

```
Caller -> slack-msg history --channel CH (no token) -> bot token required -> exit 1
```

## Preconditions

- No `--config`, no `SLACK_BOT_TOKEN`.

## Steps

1. Clear Slack env vars.
2. Provide channel only.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ClearSlackEnv = true
	req.SlackAPIURL = ""
	return nil
}
```
