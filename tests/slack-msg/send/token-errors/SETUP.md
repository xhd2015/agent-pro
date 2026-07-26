# Scenario

**Feature**: reject missing bot token when no env/config supplies one

```
Caller -> slack-msg send --channel CH MESSAGE (no token) -> stderr bot token required -> exit 1
```

## Preconditions

- No `--config`, no `SLACK_BOT_TOKEN`.

## Steps

1. Clear Slack env vars.
2. Provide `--channel` and MESSAGE only.

## Context

- Token resolution: CLI → `SLACK_BOT_TOKEN` → config `botToken`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ClearSlackEnv = true
	req.SlackAPIURL = ""
	return nil
}
```
