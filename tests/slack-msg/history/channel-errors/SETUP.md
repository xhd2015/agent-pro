# Scenario

**Feature**: reject missing channel for history

```
Caller -> slack-msg history --token TOK (no channel) -> channel required -> exit 1
```

## Preconditions

- No `--config`, no `SLACK_CHANNEL`.

## Steps

1. Clear Slack env vars.
2. Provide token only.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ClearSlackEnv = true
	req.SlackAPIURL = ""
	return nil
}
```
