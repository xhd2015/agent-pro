# Scenario

**Feature**: reject missing tokens before Socket Mode connect

```
Caller -> slack-listen listen (missing bot or app token) -> stderr error -> exit 1
```

## Preconditions

- No `--config`, no `SLACK_BOT_TOKEN` / `SLACK_APP_TOKEN`.

## Steps

1. Clear Slack env vars.
2. Leaf supplies partial token flags.

## Context

- Token resolution: CLI → env → config `botToken`/`appToken`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ClearSlackEnv = true
	req.SlackAPIURL = ""
	return nil
}
```