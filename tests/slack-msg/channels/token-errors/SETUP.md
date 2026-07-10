# Scenario

**Feature**: reject missing bot token for channels

```
Caller -> slack-msg channels list (no token) -> bot token required -> exit 1
```

## Preconditions

- No `--config`, no `SLACK_BOT_TOKEN`.

## Steps

1. Clear Slack env vars.
2. Leaf runs channels list without token.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ClearSlackEnv = true
	req.SlackAPIURL = ""
	return nil
}
```
