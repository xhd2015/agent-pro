# Scenario

**Feature**: reject missing bot token for auth status default mode

```
Caller -> slack-msg auth status (no token) -> stderr bot token required -> exit 1
```

## Preconditions

- No `--config`, no `SLACK_BOT_TOKEN`, no `--token`.

## Steps

1. Clear Slack env vars.
2. Provide only `auth status`.

## Context

- Token resolution: CLI → `SLACK_BOT_TOKEN` → config `botToken`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ClearSlackEnv = true
	req.SlackAPIURL = ""
	return nil
}
```
